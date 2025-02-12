# -*- coding: utf-8 -*-
"""
TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at https://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
from django.db.models import F, Q
from django.forms import model_to_dict

from backend.db_meta.enums import ClusterType
from backend.db_meta.models import Cluster, Machine, ProxyInstance, StorageInstance
from backend.db_services.quick_search.constants import FilterType, ResourceType
from backend.flow.models import FlowTree
from backend.ticket.models import Ticket


class QSearchHandler(object):
    def __init__(self, bk_biz_ids=None, db_types=None, resource_types=None, filter_type=None):
        self.bk_biz_ids = bk_biz_ids
        self.db_types = db_types
        self.resource_types = resource_types
        self.filter_type = filter_type
        self.limit = 10

        # db_type -> cluster_type
        self.cluster_types = []
        if self.db_types:
            for db_type in self.db_types:
                self.cluster_types.extend(ClusterType.db_type_to_cluster_type(db_type))

    def search(self, keyword):
        result = {}
        target_resource_types = self.resource_types or ResourceType.get_values()
        for target_resource_type in target_resource_types:
            filter_func = getattr(self, f"filter_{target_resource_type}", None)
            if callable(filter_func):
                result[target_resource_type] = filter_func(keyword)

        return result

    def common_filter(self, objs, return_type="list", fields=None, limit=None):
        """
        return_type: list | objects
        """
        if self.bk_biz_ids:
            objs = objs.filter(bk_biz_id__in=self.bk_biz_ids)
        if self.db_types:
            objs = objs.filter(cluster_type__in=self.cluster_types)

        if return_type == "objects":
            return objs

        fields = fields or []
        limit = limit or self.limit
        return list(objs[:limit].values(*fields))

    def filter_cluster_name(self, keyword):
        """过滤集群名"""

        qs = Q(name=keyword) if self.filter_type == FilterType.EXACT.value else Q(name__contains=keyword)
        objs = Cluster.objects.filter(qs)
        return self.common_filter(objs)

    def filter_cluster_domain(self, keyword):
        """过滤集群域名"""

        qs = (
            Q(immute_domain=keyword)
            if self.filter_type == FilterType.EXACT.value
            else Q(immute_domain__contains=keyword)
        )

        objs = Cluster.objects.filter(qs)
        return self.common_filter(objs)

    def filter_instance(self, keyword):
        """过滤实例"""

        if self.filter_type == FilterType.EXACT.value:
            qs = Q(machine__ip=keyword) | Q(name=keyword)
        else:
            qs = Q(machine__ip__contains=keyword) | Q(name__contains=keyword)

        if self.bk_biz_ids:
            qs = Q(bk_biz_id__in=self.bk_biz_ids) & qs

        if self.db_types:
            qs = Q(cluster_type__in=self.cluster_types) & qs

        objs = (
            StorageInstance.objects.filter(qs)
            .annotate(
                role=F("instance_role"),
                cluster_id=F("cluster__id"),
                cluster_domain=F("cluster__immute_domain"),
                ip=F("machine__ip"),
            )
            .union(
                ProxyInstance.objects.filter(qs).annotate(
                    role=F("access_layer"),
                    cluster_id=F("cluster__id"),
                    cluster_domain=F("cluster__immute_domain"),
                    ip=F("machine__ip"),
                )
            )
        )

        return objs[: self.limit].values(
            "id",
            "name",
            "bk_biz_id",
            "cluster_id",
            "cluster_domain",
            "cluster_type",
            "role",
            "ip",
            "port",
            "machine_type",
            "machine_id",
            "status",
            "phase",
        )

    def filter_task(self, keyword):
        """过滤任务"""

        if self.filter_type == FilterType.EXACT.value:
            qs = Q(root_id=keyword)
        else:
            qs = Q(root_id__contains=keyword)

        objs = FlowTree.objects.filter(qs)

        if self.bk_biz_ids:
            objs = objs.filter(bk_biz_id__in=self.bk_biz_ids)

        # TODO: db类型任务的过滤
        return list(objs[: self.limit].values("uid", "bk_biz_id", "ticket_type", "root_id", "status", "created_by"))

    def filter_machine(self, keyword):
        """过滤主机"""

        if self.filter_type == FilterType.EXACT.value:
            qs = Q(ip=keyword)

            try:
                # fix bk_host_id expected a number but got string
                keyword_int = int(keyword)
                qs = qs | Q(bk_host_id=keyword_int)
            except ValueError:
                pass
        else:
            qs = Q(ip__contains=keyword)

        if self.bk_biz_ids:
            qs = qs & Q(bk_biz_id__in=self.bk_biz_ids)

        if self.db_types:
            qs = qs & Q(cluster_type__in=self.cluster_types)

        objs = Machine.objects.filter(qs).prefetch_related(
            "storageinstance_set__cluster", "proxyinstance_set__cluster"
        )

        # 解析cluster
        machines = []
        for obj in objs[: self.limit]:
            machine = model_to_dict(
                obj, ["bk_biz_id", "bk_host_id", "ip", "cluster_type", "spec_id", "bk_cloud_id", "bk_city"]
            )

            # 兼容实例未绑定集群的情况
            inst = obj.storageinstance_set.first() or obj.proxyinstance_set.first()
            cluster = inst.cluster.first() if inst and inst.cluster.exists() else None
            cluster_info = {"cluster_id": None, "cluster_domain": None}
            if cluster:
                cluster_info.update({"cluster_id": cluster.id, "cluster_domain": cluster.immute_domain})

            machine.update(cluster_info)
            machines.append(machine)

        return machines

    def filter_ticket(self, keyword):
        """过滤单据，单号为递增数字，采用startswith过滤"""

        try:
            ticket_id = int(keyword)
            qs = Q(id=ticket_id) if self.filter_type == FilterType.EXACT.value else Q(id__startswith=keyword)
            objs = Ticket.objects.filter(qs).order_by("id")
            return self.common_filter(
                objs,
                fields=[
                    "id",
                    "creator",
                    "create_at",
                    "bk_biz_id",
                    "ticket_type",
                    "group",
                    "status",
                    "is_reviewed",
                ],
                limit=5,
            )
        except ValueError:
            return []
