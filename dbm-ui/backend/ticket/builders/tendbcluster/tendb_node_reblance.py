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

from django.utils.translation import gettext_lazy as _
from rest_framework import serializers

from backend.db_meta.models import Cluster
from backend.db_services.dbbase.constants import IpSource
from backend.flow.engine.controller.spider import SpiderController
from backend.ticket import builders
from backend.ticket.builders.tendbcluster.base import BaseTendbTicketFlowBuilder, TendbBaseOperateDetailSerializer
from backend.ticket.constants import TicketType, TriggerChecksumType


class TendbNodeRebalanceDetailSerializer(TendbBaseOperateDetailSerializer):
    class NodeRebalanceItemSerializer(serializers.Serializer):
        cluster_id = serializers.IntegerField(help_text=_("集群ID"))
        bk_cloud_id = serializers.IntegerField(help_text=_("云区域ID"))
        cluster_shard_num = serializers.IntegerField(help_text=_("集群分片数"))
        remote_shard_num = serializers.IntegerField(help_text=_("单机分片数"))
        resource_spec = serializers.JSONField(help_text=_("规格要求"))

    infos = serializers.ListSerializer(help_text=_("集群扩缩容信息"), child=NodeRebalanceItemSerializer())
    ip_source = serializers.ChoiceField(
        help_text=_("主机来源"), choices=IpSource.get_choices(), default=IpSource.RESOURCE_POOL.value
    )
    need_checksum = serializers.BooleanField(help_text=_("执行前是否需要数据校验"))
    trigger_checksum_type = serializers.ChoiceField(help_text=_("数据校验触发类型"), choices=TriggerChecksumType.get_choices())
    trigger_checksum_time = serializers.CharField(help_text=_("数据校验 触发时间"))

    def validate(self, attrs):
        super().validate(attrs)

        # 校验集群分片数匹配单机分片数
        for info in attrs["infos"]:
            super().validate_cluster_shard_num(info)

        return attrs


class TendbNodeRebalanceFlowParamBuilderBuilder(builders.FlowParamBuilder):
    controller = SpiderController.tendb_cluster_remote_rebalance

    def format_ticket_data(self):
        cluster_ids = [info["cluster_id"] for info in self.ticket_data["infos"]]
        cluster__module_id = {
            cluster.id: cluster.db_module_id for cluster in Cluster.objects.filter(id__in=cluster_ids)
        }
        for info in self.ticket_data["infos"]:
            info["db_module_id"] = cluster__module_id[info["cluster_id"]]


class TendbNodeRebalanceResourceParamBuilder(builders.ResourceApplyParamBuilder):
    def format(self):
        self.patch_info_affinity_location(roles=["backend_group"])

    def post_callback(self):
        next_flow = self.ticket.next_flow()
        infos = next_flow.details["ticket_data"]["infos"]
        for info in infos:
            info["resource_spec"]["remote"] = info["resource_spec"].pop("master")
            info["resource_spec"].pop("slave")
            info["remote_group"] = info.pop("backend_group")

        next_flow.save(update_fields=["details"])


@builders.BuilderFactory.register(TicketType.TENDBCLUSTER_NODE_REBALANCE, is_apply=True)
class TendbMNTApplyFlowBuilder(BaseTendbTicketFlowBuilder):
    serializer = TendbNodeRebalanceDetailSerializer
    inner_flow_builder = TendbNodeRebalanceFlowParamBuilderBuilder
    resource_batch_apply_builder = TendbNodeRebalanceResourceParamBuilder
    inner_flow_name = _("TendbCluster 集群容量变更")
