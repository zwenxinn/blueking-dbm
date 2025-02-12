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

from dataclasses import asdict
from typing import Dict, Optional

from django.utils.translation import ugettext as _

from backend import env
from backend.components.dbresource.client import DBResourceApi
from backend.configuration.constants import SystemSettingsEnum
from backend.configuration.models import SystemSettings
from backend.flow.engine.bamboo.scene.common.builder import Builder
from backend.flow.plugins.components.collections.common.external_service import ExternalServiceComponent
from backend.flow.plugins.components.collections.common.sa_idle_check import CheckMachineIdleComponent
from backend.flow.plugins.components.collections.common.sa_init import SaInitComponent
from backend.flow.plugins.components.collections.common.transfer_host_service import TransferHostServiceComponent
from backend.flow.utils.mysql.mysql_act_dataclass import InitCheckForResourceKwargs


class ImportResourceInitStepFlow(object):
    """
    机器初始化步骤
    """

    def __init__(self, root_id: str, data: Optional[Dict]) -> None:
        self.root_id = root_id
        self.data = data

    def machine_init_flow(self):
        p = Builder(root_id=self.root_id, data=self.data)
        ip_list = self.data["hosts"]
        bk_biz_id = self.data["bk_biz_id"]
        # 先执行空闲检查
        if env.SA_CHECK_TEMPLATE_ID:
            p.add_act(
                act_name=_("执行sa空闲检查"),
                act_component_code=CheckMachineIdleComponent.code,
                kwargs=asdict(InitCheckForResourceKwargs(ips=[host["ip"] for host in ip_list], bk_biz_id=bk_biz_id)),
            )

        # 在执行sa初始化
        if env.SA_INIT_TEMPLATE_ID:
            # 执行sa初始化
            p.add_act(
                act_name=_("执行sa初始化"),
                act_component_code=SaInitComponent.code,
                kwargs={"ips": [host["ip"] for host in ip_list], "bk_biz_id": bk_biz_id},
            )

        # 调用资源导入接口
        p.add_act(
            act_name=_("资源池导入"),
            act_component_code=ExternalServiceComponent.code,
            kwargs={
                "params": self.data,
                "api_import_path": DBResourceApi.__module__,
                "api_import_module": "DBResourceApi",
                "api_call_func": "resource_import",
            },
        )

        # 转移模块到资源池空闲机
        p.add_act(
            act_name=_("主机转移至资源池空闲模块"),
            act_component_code=TransferHostServiceComponent.code,
            kwargs={
                "bk_biz_id": env.DBA_APP_BK_BIZ_ID,
                "bk_module_ids": [
                    SystemSettings.get_setting_value(key=SystemSettingsEnum.MANAGE_TOPO.value)["resource_module_id"]
                ],
                "bk_host_ids": [host["host_id"] for host in ip_list],
            },
        )

        p.run_pipeline()
