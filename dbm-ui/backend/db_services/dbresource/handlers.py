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
import itertools
import math
from typing import Any, Dict, List

from django.forms import model_to_dict
from django.utils.translation import ugettext as _

from backend.components.dbresource.client import DBResourceApi
from backend.db_meta.enums import ClusterType
from backend.db_meta.models import Spec
from backend.db_services.dbresource.exceptions import SpecOperateException


class ClusterSpecFilter(object):
    """集群规格的过滤器"""

    def __init__(self, capacity, future_capacity, spec_cluster_type, spec_machine_type, qps=None, shard_num=0):
        # 用户的当前容量，期望容量，期望qps范围和分片数(可选)
        self.capacity: int = capacity
        self.future_capacity: int = future_capacity
        self.qps: Dict = qps
        self.filter_shard_num = shard_num
        # 当前集群的筛选规格
        self.specs: List[Dict[str, Any]] = [
            {**model_to_dict(spec), "capacity": spec.capacity}
            for spec in Spec.objects.filter(spec_machine_type=spec_machine_type, spec_cluster_type=spec_cluster_type)
        ]

    def calc_machine_pair(self):
        """计算每种规格所需的机器组数和集群总容量: 目标容量 / 规格容量"""
        for spec in self.specs:
            spec["machine_pair"] = math.ceil(self.capacity / spec["capacity"])
            # 集群容量：机器组数 * 规格容量；集群qps：机器组数 * 规格qps的min
            spec["cluster_capacity"] = spec["machine_pair"] * spec["capacity"]
            if self.qps:
                spec["cluster_qps"] = spec["machine_pair"] * spec["qps"]["min"]

    def calc_cluster_shard_num(self):
        """计算每种规格的分片数, 根据不同的集群计算方式也不同"""
        raise NotImplementedError()

    def _qps_check(self, user_qps_range, spec_qps_range):
        """默认判断规则：当前qps与用户需要存在交集"""
        if user_qps_range["min"] > spec_qps_range["max"] or user_qps_range["max"] < spec_qps_range["min"]:
            return False

        return True

    def system_filter(self):
        """系统自带的过滤：qps和分片数"""
        valid_specs: List[Dict[str, Any]] = []
        for spec in self.specs:
            if self.filter_shard_num and spec["cluster_shard_num"] != self.filter_shard_num:
                continue

            if self.qps:
                qps_range = {
                    "min": spec["machine_pair"] * spec["qps"]["min"],
                    "max": spec["machine_pair"] * spec["qps"]["max"],
                }
                if not self._qps_check(self.qps, qps_range):
                    continue

            valid_specs.append(spec)

        self.specs = valid_specs

    def custom_filter(self):
        """自定义过滤规则"""
        pass

    def get_target_specs(self):
        self.calc_machine_pair()
        self.calc_cluster_shard_num()
        self.system_filter()
        self.custom_filter()
        return self.specs


class TenDBClusterSpecFilter(ClusterSpecFilter):
    """TendbCluster集群规格的过滤器"""

    def calc_cluster_shard_num(self):
        for spec in self.specs:
            # 一定要保证集群总分片数是机器组数的整数倍，因此单机分片数要上取整
            cluster_shard_num = math.ceil(self.future_capacity / spec["capacity"])
            single_machine_shard_num = math.ceil(cluster_shard_num / spec["machine_pair"])
            spec["cluster_shard_num"] = single_machine_shard_num * spec["machine_pair"]


class RedisSpecFilter(ClusterSpecFilter):
    """Redis规格过滤器基类"""

    def _qps_check(self, user_qps_range, spec_qps_range):
        # redis不需要qps校验
        return True

    def custom_filter(self):
        """对规格方案进行排序,根据目标容量,以及未来容量来决定排序方式"""
        if self.future_capacity >= self.capacity:
            self.specs.sort(key=lambda x: (-x["cluster_shard_num"], -x["capacity"], x["machine_pair"]))
        else:
            self.specs.sort(key=lambda x: (x["cluster_shard_num"], -x["capacity"], x["machine_pair"]))


class TendisPlusSpecFilter(RedisSpecFilter):
    """TendisPlus集群规格过滤器"""

    # 最佳容量管理大小 300G
    OPTIMAL_MANAGE_CAPACITY = 300

    def calc_machine_pair(self):
        """计算每种规格所需的机器组数，TendisPlus至少需要三组"""
        for spec in self.specs:
            spec["machine_pair"] = max(math.ceil(self.capacity / spec["capacity"]), 3)
            spec["cluster_capacity"] = spec["machine_pair"] * spec["capacity"]

    def calc_cluster_shard_num(self):
        for spec in self.specs:
            spec["cluster_shard_num"] = max(3, math.ceil(self.capacity / self.OPTIMAL_MANAGE_CAPACITY))
            # 将分片数上取整为机器组数的倍数
            spec["cluster_shard_num"] = (
                math.ceil(spec["cluster_shard_num"] / spec["machine_pair"]) * spec["machine_pair"]
            )

    def custom_filter(self):
        super().custom_filter()


class TendisSSDSpecFilter(RedisSpecFilter):
    """TendisSSD集群规格过滤器"""

    # 单实例最大容量 50G
    SINGLE_MAX_CAPACITY = 50
    MACHINE_PAIR_SORT = True

    def calc_cluster_shard_num(self):
        for spec in self.specs:
            # 计算单机分片数，容量/50-取整为最接近的偶数
            single_machine_shard_num = int(spec["capacity"] / self.SINGLE_MAX_CAPACITY)
            single_machine_shard_num = max(single_machine_shard_num + (single_machine_shard_num & 1), 2)
            spec["cluster_shard_num"] = single_machine_shard_num * spec["machine_pair"]

    def custom_filter(self):
        super().custom_filter()


class TendisCacheSpecFilter(RedisSpecFilter):
    """TendisCache集群规格过滤器"""

    # 单个实例建议的容量规格
    BASE_SINGLE_CAPCITY = 6
    # 支持简单阔缩容倍数（非DTS方式）
    SCALE_MULITPLE = 4

    def calc_cluster_shard_num(self):
        valid_specs: List[Dict[str, Any]] = []
        max_capcity = self.capacity
        if self.future_capacity > self.capacity:
            max_capcity = min(self.future_capacity, self.capacity * self.SCALE_MULITPLE / 2)
        # 先进行排序
        self.specs.sort(key=lambda x: (x["capacity"]))

        # 选取合适的规格
        spec_idx, instance_cap, spec_cnt, avaiable_specs = 0, self.BASE_SINGLE_CAPCITY, len(self.specs), []
        for spec in self.specs:
            if self.capacity <= spec["capacity"]:
                avaiable_specs.append(spec)
                if spec_idx >= 1:
                    avaiable_specs.append(self.specs[spec_idx - 1])
                if spec_idx >= 3:
                    avaiable_specs.append(self.specs[spec_idx - 2])
                break
            spec_idx += 1

        if self.capacity > self.specs[spec_cnt - 1]["capacity"] * self.SCALE_MULITPLE:
            instance_cap = self.BASE_SINGLE_CAPCITY * self.SCALE_MULITPLE

        if self.capacity > self.specs[spec_cnt - 1]["capacity"]:
            avaiable_specs.append(self.specs[spec_cnt - 1])
            if spec_cnt > 2:
                avaiable_specs.append(self.specs[spec_cnt - 2])

        for spec_new in avaiable_specs:
            # 一定要保证集群总分片数是机器组数的整数倍，并且单机分片数需要取整，取偶
            cluster_shard_num = math.ceil(max_capcity / instance_cap)
            single_machine_shard_num = math.ceil(cluster_shard_num / spec_new["machine_pair"])
            single_machine_shard_num = max(single_machine_shard_num + (single_machine_shard_num & 1), 2)
            spec_new["cluster_shard_num"] = max(single_machine_shard_num * spec_new["machine_pair"], 4)

            valid_specs.append(spec_new)

        self.specs = valid_specs

    def custom_filter(self):
        super().custom_filter()


class ResourceHandler(object):
    """资源池接口的处理函数"""

    @classmethod
    def spec_resource_count(cls, bk_biz_id: int, bk_cloud_id: int, spec_ids: List[int]):
        specs = Spec.objects.filter(spec_id__in=spec_ids)
        # 获取resource_type
        spec_cluster_type = list(set(specs.values_list("spec_cluster_type", flat=True)))
        if len(spec_cluster_type) > 1:
            raise SpecOperateException(_("请保证请求的规格类型一致"))
        resource_type = ClusterType.cluster_type_to_db_type(spec_cluster_type[0])
        # 构造申请参数
        spec_count_details = [
            spec.get_group_apply_params(group_mark=str(spec.spec_id), count=1, group_count=1, bk_cloud_id=bk_cloud_id)
            for spec in specs
        ]
        spec_count_details = list(itertools.chain(*spec_count_details))
        spec_count_params = {
            "for_biz_id": bk_biz_id,
            "resource_type": resource_type,
            "bk_cloud_id": bk_cloud_id,
            "details": spec_count_details,
        }
        # 获取规格的预估数量，注意剔除分组后缀
        spec_apply_count = DBResourceApi.apply_count(params=spec_count_params)
        spec_apply_count = {k.split("_")[0]: v for k, v in spec_apply_count.items()}
        return spec_apply_count
