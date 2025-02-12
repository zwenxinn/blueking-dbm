package atomredis

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"dbm-services/redis/db-tools/dbactuator/pkg/common"
	"dbm-services/redis/db-tools/dbactuator/pkg/consts"
	"dbm-services/redis/db-tools/dbactuator/pkg/jobruntime"
	"dbm-services/redis/db-tools/dbactuator/pkg/util"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

// ConfServerItem servers配置项
type ConfServerItem struct {
	BkBizID         string            `json:"bk_biz_id" yaml:"bk_biz_id" validate:"required"`
	BkCloudID       int64             `json:"bk_cloud_id" yaml:"bk_cloud_id"`
	App             string            `json:"app" yaml:"app" validate:"required"`
	AppName         string            `json:"app_name" yaml:"app_name" validate:"required"`
	ClusterDomain   string            `json:"cluster_domain" yaml:"cluster_domain" validate:"required"`
	ClusterName     string            `json:"cluster_name" yaml:"cluster_name" validate:"required"`
	ClusterType     string            `json:"cluster_type" yaml:"cluster_type" validate:"required"`
	MetaRole        string            `json:"meta_role" yaml:"meta_role" validate:"required"`
	ServerIP        string            `json:"server_ip" yaml:"server_ip" validate:"required"`
	ServerPorts     []int             `json:"server_ports" yaml:"server_ports" validate:"required"`
	ServerShards    map[string]string `json:"server_shards" yaml:"server_shards"`
	CacheBackupMode string            `json:"cache_backup_mode" yaml:"cache_backup_mode"` // aof or rdb
	Shard           string            `json:"shard" yaml:"shard"`
}

// BkDbmonInstallParams 安装参数
type BkDbmonInstallParams struct {
	BkDbmonPkg               common.MediaPkg        `json:"bkdbmonpkg" validate:"required"`
	DbToolsPkg               common.DbToolsMediaPkg `json:"dbtoolspkg" validate:"required"`
	AgentAddress             string                 `json:"agent_address" validate:"required"`
	BeatPath                 string                 `json:"beat_path" validate:"required"`
	BackupClientStrorageType string                 `json:"backup_client_storage_type"`
	RedisFullBackup          map[string]interface{} `json:"redis_fullbackup" validate:"required"`
	RedisBinlogBackup        map[string]interface{} `json:"redis_binlogbackup" validate:"required"`
	RedisHeartbeat           map[string]interface{} `json:"redis_heartbeat" validate:"required"`
	RedisMonitor             map[string]interface{} `json:"redis_monitor" validate:"required"`
	RedisKeyLifecyckle       map[string]interface{} `json:"redis_keylife" mapstructure:"redis_keylife"`
	Servers                  []ConfServerItem       `json:"servers" yaml:"servers" validate:"required"`
}

// BkDbmonInstall bk-dbmon安装任务
type BkDbmonInstall struct {
	runtime           *jobruntime.JobGenericRuntime
	params            BkDbmonInstallParams
	bkDbmonBinUpdated bool // bk-dbmon介质是否被更新
	confFileUpdated   bool //  配置文件是否被更新
	isStopped         bool // 是否是停止 bkdbmon
}

// 无实际作用,仅确保实现了 jobruntime.JobRunner 接口
var _ jobruntime.JobRunner = (*RedisInstall)(nil)

// NewBkDbmonInstall new
func NewBkDbmonInstall() jobruntime.JobRunner {
	return &BkDbmonInstall{}
}

// Init 初始化
func (job *BkDbmonInstall) Init(m *jobruntime.JobGenericRuntime) error {
	job.runtime = m
	d := json.NewDecoder(strings.NewReader(job.runtime.PayloadDecoded))
	d.UseNumber()
	err := d.Decode(&job.params)
	if err != nil {
		job.runtime.Logger.Error(fmt.Sprintf("json.Decode failed,err:%+v", err))
		return err
	}
	// 参数有效性检查
	validate := validator.New()
	err = validate.Struct(job.params)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:%v,params:%+v",
				err, job.params)
			return err
		}
		for _, err := range err.(validator.ValidationErrors) {
			job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:%v,params:%+v",
				err, job.params)
			return err
		}
	}
	for _, svrItem := range job.params.Servers {
		if len(svrItem.ServerPorts) > 0 {
			if svrItem.ServerIP == "" {
				job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:ServerIP is empty")
				return fmt.Errorf("ServerIP is empty")
			}
			if svrItem.ClusterName == "" {
				job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:ClusterName is empty")
				return fmt.Errorf("ClusterName is empty")
			}
			if svrItem.ClusterDomain == "" {
				job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:ClusterDomain is empty")
				return fmt.Errorf("ClusterDomain is empty")
			}
			if svrItem.ClusterType == "" {
				job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:ClusterType is empty")
				return fmt.Errorf("ClusterType is empty")
			}
			if svrItem.ClusterType == "" {
				job.runtime.Logger.Error("BkDbmonInstall Init params validate failed,err:ClusterType is empty")
				return fmt.Errorf("ClusterType is empty")
			}
		}
	}
	return nil
}

// Name 原子任务名
func (job *BkDbmonInstall) Name() string {
	return "bkdbmon_install"
}

// Run 执行
func (job *BkDbmonInstall) Run() (err error) {
	err = job.checkIsStopped()
	if err != nil {
		return
	}
	err = job.UntarMedia()
	if err != nil {
		return
	}
	err = job.GenerateConfigFile()
	if err != nil {
		return
	}
	err = job.stopDbmonWhenNoServers()
	if err != nil {
		return
	}
	if job.isStopped {
		return
	}
	if !job.bkDbmonBinUpdated && !job.confFileUpdated {
		job.runtime.Logger.Info("bk-dbmon media,configfile both not updated")
	} else {
		err = job.StopBkDbmon()
	}
	if err != nil {
		return
	}
	err = job.StartBkDbmon()
	if err != nil {
		return
	}
	err = job.newExporterConfig()
	if err != nil {
		return
	}

	return
}

func (job *BkDbmonInstall) checkIsStopped() (err error) {
	if len(job.params.Servers) == 0 {
		job.isStopped = true
		return
	}
	job.isStopped = true
	for _, svrItem := range job.params.Servers {
		if len(svrItem.ServerPorts) > 0 {
			// 存在实例
			job.isStopped = false
			break
		}
	}
	return
}

// stopDbmonWhenNoServers servers为空,没有任何需要监控的 redis-servers/proxy
// 直接stop bkdbmon即可
func (job *BkDbmonInstall) stopDbmonWhenNoServers() (err error) {
	if job.isStopped {
		err = job.StopBkDbmon()
		return
	}
	return
}

// UntarMedia 解压介质
// 如果/home/mysql/bk-dbmon/bk-dbmon 存在,且版本正确,则不解压
// 否则解压最新bk-dbmon,并修改 /home/mysql/bk-dbmon 的指向;
func (job *BkDbmonInstall) UntarMedia() (err error) {
	if job.isStopped {
		// 如果是停止 bkdbmon,则不需要解压介质
		return nil
	}
	var remoteVersion, localVersion string
	err = job.params.BkDbmonPkg.Check()
	if err != nil {
		job.runtime.Logger.Error("UntarMedia check failed,err:%v,skip...", err)
		// return
	}
	defer util.LocalDirChownMysql(consts.BkDbmonPath + "/")
	verReg := regexp.MustCompile(`bk-dbmon-(v\d+.\d+).tar.gz`)
	l01 := verReg.FindStringSubmatch(job.params.BkDbmonPkg.Pkg)
	if len(l01) != 2 {
		err = fmt.Errorf("%s format not correct? for example bk-dbmon-v0.1.tar.gz",
			job.params.BkDbmonPkg.Pkg)
		job.runtime.Logger.Error(err.Error())
		return
	}
	remoteVersion = l01[1]
	if util.FileExists(consts.BkDbmonBin) {
		cmd := fmt.Sprintf("%s -v |awk '{print $2}'", consts.BkDbmonBin)
		localVersion, err = util.RunBashCmd(cmd, "", nil, 1*time.Minute)
		if err != nil {
			return
		}
		localVersion = strings.TrimSpace(localVersion)
	}
	if remoteVersion != "" && remoteVersion == localVersion {
		// 如果本地版本和远程版本一致,则无需更新
		job.runtime.Logger.Info("本地bk-dbmon版本%s 与 目标bk-dbmon版本%s 一致,无需更新本地bk-dbmon版本", localVersion, remoteVersion)
		return
	}
	job.bkDbmonBinUpdated = true
	err = job.StopBkDbmon()
	if err != nil {
		return
	}
	err = job.RemoveBkDbmon()
	if err != nil {
		return
	}

	// 解压新版本
	pkgAbsPath := job.params.BkDbmonPkg.GetAbsolutePath()
	pkgBasename := job.params.BkDbmonPkg.GePkgBaseName()
	bakDir := filepath.Join(consts.GetRedisBackupDir(), "dbbak")
	tarCmd := fmt.Sprintf(" tar -zxf %s -C %s", pkgAbsPath, bakDir)
	job.runtime.Logger.Info(tarCmd)
	_, err = util.RunBashCmd(tarCmd, "", nil, 1*time.Minute)
	if err != nil {
		return
	}
	bkDbmonRealDir := filepath.Join(bakDir, pkgBasename)
	if !util.FileExists(bkDbmonRealDir) {
		err = fmt.Errorf("untar %s success but %s not exists", pkgAbsPath, bkDbmonRealDir)
		job.runtime.Logger.Error(err.Error())
		return
	}
	// 创建软链接
	var stat fs.FileInfo
	var link string
	stat, err = os.Stat(consts.BkDbmonPath)
	if err == nil && stat.Mode()&os.ModeSymlink != 0 {
		link, err = os.Readlink(consts.BkDbmonPath)
		if err == nil && link == bkDbmonRealDir {
			// 软链接已经存在,且指向正确,无需再创建
			job.runtime.Logger.Info("软链接%s已经存在,且指向正确,无需再创建", consts.BkDbmonPath)
			return
		}
	}
	err = os.Symlink(bkDbmonRealDir, consts.BkDbmonPath)
	if err != nil {
		err = fmt.Errorf("os.Symlink failed,err:%v,dir:%s,softLink:%s", err, bkDbmonRealDir, consts.BkDbmonPath)
		job.runtime.Logger.Error(err.Error())
		return
	}
	util.LocalDirChownMysql(bkDbmonRealDir)
	err = job.params.DbToolsPkg.Install()
	if err != nil {
		return err
	}
	return nil
}

// StopBkDbmon stop local bk-dbmon
func (job *BkDbmonInstall) StopBkDbmon() (err error) {
	err = util.StopBkDbmon()
	return
}

// StartBkDbmon start local bk-dbmon
func (job *BkDbmonInstall) StartBkDbmon() (err error) {
	return util.StartBkDbmon()
}

// RemoveBkDbmon remove local bk-dbmon
func (job *BkDbmonInstall) RemoveBkDbmon() (err error) {
	if !util.FileExists(consts.BkDbmonPath) {
		return
	}
	job.runtime.Logger.Info("RemoveBkDbmon %s exists,start remove it", consts.BkDbmonPath)
	var realDir string
	if util.FileExists(consts.BkDbmonPath) {
		realDir, err = filepath.EvalSymlinks(consts.BkDbmonPath)
		if err != nil {
			err = fmt.Errorf("filepath.EvalSymlinks failed,err:%v,bkDbmon:%s", err, consts.BkDbmonPath)
			job.runtime.Logger.Error(err.Error())
			return
		}
		rmCmd := fmt.Sprintf("rm -rf %s", consts.BkDbmonPath)
		job.runtime.Logger.Info(rmCmd)
		util.RunBashCmd(rmCmd, "", nil, 1*time.Minute)
	}
	if realDir != "" && util.FileExists(realDir) {
		rmCmd := fmt.Sprintf("rm -rf %s", realDir)
		job.runtime.Logger.Info(rmCmd)
		util.RunBashCmd(rmCmd, "", nil, 1*time.Minute)
	}
	return
}

// bkDbmonConf 生成bk-dbmon配置
type bkDbmonConf struct {
	ReportSaveDir            string                 `json:"report_save_dir" yaml:"report_save_dir"`
	ReportLeftDay            int                    `json:"report_left_day" yaml:"report_left_day"`
	HTTPAddress              string                 `json:"http_address" yaml:"http_address"`
	AgentAddress             string                 `json:"agent_address" yaml:"agent_address"`
	BeatPath                 string                 `json:"beat_path" yaml:"beat_path"`
	BackupClientStrorageType string                 `json:"backup_client_storage_type" yaml:"backup_client_storage_type"`
	RedisFullBackup          map[string]interface{} `json:"redis_fullbackup" yaml:"redis_fullbackup"`
	RedisBinlogBackup        map[string]interface{} `json:"redis_binlogbackup" yaml:"redis_binlogbackup"`
	RedisHeartbeat           map[string]interface{} `json:"redis_heartbeat" yaml:"redis_heartbeat"`
	RedisMonitor             map[string]interface{} `json:"redis_monitor" yaml:"redis_monitor"`
	RedisKeyLifecyckle       map[string]interface{} `json:"redis_keylife" yaml:"redis_keylife"`
	Servers                  []ConfServerItem       `json:"servers" yaml:"servers"`
}

// ToString string
func (conf *bkDbmonConf) ToString() string {
	tmp, _ := json.Marshal(conf)
	return string(tmp)
}

// GenerateConfigFile 生成bk-dbmon的配置
func (job *BkDbmonInstall) GenerateConfigFile() (err error) {
	var yamlData []byte
	var confMd5, tempMd5 string
	var notUpdateConf bool = false
	confData := &bkDbmonConf{
		ReportSaveDir:            consts.DbaReportSaveDir,
		ReportLeftDay:            consts.RedisReportLeftDay,
		HTTPAddress:              consts.BkDbmonHTTPAddress,
		AgentAddress:             job.params.AgentAddress,
		BeatPath:                 job.params.BeatPath,
		BackupClientStrorageType: job.params.BackupClientStrorageType,
		RedisFullBackup:          job.params.RedisFullBackup,
		RedisBinlogBackup:        job.params.RedisBinlogBackup,
		RedisHeartbeat:           job.params.RedisHeartbeat,
		RedisMonitor:             job.params.RedisMonitor,
		RedisKeyLifecyckle:       job.params.RedisKeyLifecyckle,
		Servers:                  job.params.Servers,
	}

	yamlData, err = yaml.Marshal(confData)
	if err != nil {
		err = fmt.Errorf("yaml.Marshal fail,err:%v", err)
		job.runtime.Logger.Info(err.Error())
		return
	}
	tempFile := consts.BkDbmonConfFile + "_temp"
	err = ioutil.WriteFile(tempFile, yamlData, 0755)
	if err != nil {
		err = fmt.Errorf("ioutil.WriteFile fail,err:%v", err)
		job.runtime.Logger.Info(err.Error())
		return
	}
	if util.FileExists(consts.BkDbmonConfFile) {
		confMd5, err = util.GetFileMd5(consts.BkDbmonConfFile)
		if err != nil {
			job.runtime.Logger.Error(err.Error())
			return
		}
		tempMd5, err = util.GetFileMd5(tempFile)
		if err != nil {
			job.runtime.Logger.Error(err.Error())
			return
		}
		if confMd5 == tempMd5 {
			os.Remove(tempFile)
			notUpdateConf = true
		}
	}
	if notUpdateConf {
		job.runtime.Logger.Info("config file(%s) no need update", consts.BkDbmonConfFile)
		return
	}
	job.confFileUpdated = true
	mvCmd := fmt.Sprintf("mv %s %s", tempFile, consts.BkDbmonConfFile)
	job.runtime.Logger.Info(mvCmd)
	_, err = util.RunBashCmd(mvCmd, "", nil, 1*time.Minute)
	if err != nil {
		return
	}
	util.LocalDirChownMysql(consts.BkDbmonConfFile)
	return
}

func (job *BkDbmonInstall) newExporterConfig() (err error) {
	job.runtime.Logger.Info("begin to new exporter config file")
	err = util.MkDirsIfNotExists([]string{consts.ExporterConfDir})
	if err != nil {
		job.runtime.Logger.Error("newExporterConfig mkdirIfNotExists %s failed,err:%v", consts.ExporterConfDir, err)
		return err
	}
	for _, server := range job.params.Servers {
		for _, port := range server.ServerPorts {
			err = common.CreateLocalExporterConfigFile(server.ServerIP, port, server.MetaRole, "")
			if err != nil {
				return
			}
		}
	}
	util.LocalDirChownMysql(consts.ExporterConfDir)
	return nil
}

// Retry times
func (job *BkDbmonInstall) Retry() uint {
	return 2
}

// Rollback rollback
func (job *BkDbmonInstall) Rollback() error {
	return nil
}
