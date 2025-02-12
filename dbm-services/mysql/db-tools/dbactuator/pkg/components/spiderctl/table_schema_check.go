/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package spiderctl

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"dbm-services/common/go-pubpkg/logger"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components"
	"dbm-services/mysql/db-tools/dbactuator/pkg/native"
)

// TableSchemaCheckComp TODO
type TableSchemaCheckComp struct {
	GeneralParam *components.GeneralParam `json:"general"`
	Params       TableSchemaCheckParam
	tableSchemaCheckCtx
}

// TableSchemaCheckParam TODO
type TableSchemaCheckParam struct {
	Host         string        `json:"host" validate:"required,ip"`
	Port         int           `json:"port" validate:"required,lt=65536,gte=3306"`
	CheckObjects []CheckObject `json:"check_objects" validate:"required,dive"`
}

// CheckObject TODO
type CheckObject struct {
	DbName string   `json:"dbname" validate:"required"`
	Tables []string `json:"tables" validate:"required,dive"`
}
type tableSchemaCheckCtx struct {
	tdbCtlConn *native.TdbctlDbWork
}

// TsccSchemaChecksum TODO
var TsccSchemaChecksum = `CREATE TABLE if not exists infodba_schema.tscc_schema_checksum(
	db char(64) NOT NULL,
	tbl char(64) NOT NULL,
	status char(32) NOT NULL DEFAULT "" COMMENT "检查结果,一致:ok,不一致:inconsistent",
	checksum_result json COMMENT "差异表结构信息,tdbctl checksum table 的结果",
	update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (db,tbl)
);`

// Example TODO
func (r *TableSchemaCheckComp) Example() interface{} {
	return &TableSchemaCheckComp{
		Params: TableSchemaCheckParam{
			Host: "127.0.0.1",
			Port: 26000,
			CheckObjects: []CheckObject{
				{
					DbName: "test",
					Tables: []string{"t1", "t2"},
				},
			},
		},
	}
}

// Init TODO
func (r *TableSchemaCheckComp) Init() (err error) {
	var conn *native.DbWorker
	// connection central control
	conn, err = native.InsObject{
		Host: r.Params.Host,
		Port: r.Params.Port,
		User: r.GeneralParam.RuntimeAccountParam.AdminUser,
		Pwd:  r.GeneralParam.RuntimeAccountParam.AdminPwd,
	}.Conn()
	if err != nil {
		logger.Error("connect tdbctl error: %v", err)
		return err
	}
	r.tdbCtlConn = &native.TdbctlDbWork{DbWorker: *conn}
	// init checksum table schema
	if _, err = r.tdbCtlConn.ExecMore([]string{"set tc_admin = 0;", "use infodba_schema;",
		TsccSchemaChecksum}); err != nil {
		logger.Error("init tscc_schema_checksum error: %v", err)
		return
	}
	return err
}

// Run TODO
func (r *TableSchemaCheckComp) Run() (err error) {
	for _, checkObject := range r.Params.CheckObjects {
		for _, table := range checkObject.Tables {
			// check table schema
			var result native.SchemaCheckResults
			err = r.tdbCtlConn.Queryx(&result, fmt.Sprintf("set tc_admin=1;tdbctl checksum `%s`.`%s`;", checkObject.DbName,
				table))
			if err != nil {
				logger.Error("check table schema error: %s", err.Error())
				return err
			}
			if err = r.atomUpdateCheckResult(checkObject.DbName, table, result.CheckResult()); err == nil {
				slog.Info("update checkresult ok")
			}
		}
	}
	return nil
}

func (r *TableSchemaCheckComp) atomUpdateCheckResult(db, tbl string, inconsistentItems []native.SchemaCheckResult) (
	err error) {
	r.tdbCtlConn.Exec("set tc_admin=0;")
	status := native.SchemaCheckOk
	checkResult := []byte("{}")
	if len(inconsistentItems) > 0 {
		logger.Warn("tabel %s.%s has inconsistent items", db, tbl)
		status = ""
		checkResult, err = json.Marshal(inconsistentItems)
		if err != nil {
			logger.Error("json marshal failed %s", err.Error())
			return
		}
	}
	if _, err = r.tdbCtlConn.Exec("replace into infodba_schema.tscc_schema_checksum values(?,?,?,?,?)", db,
		tbl, status,
		checkResult,
		time.Now()); err != nil {
		logger.Error("replace checksum record failed", err)
		return
	}
	return
}
