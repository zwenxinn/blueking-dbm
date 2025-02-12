// Package consts 常量
package consts

const (
	// TendisTypePredixyRedisCluster predixy + RedisCluster架构
	TendisTypePredixyRedisCluster = "PredixyRedisCluster"
	// TendisTypePredixyTendisplusCluster predixy + TendisplusCluster架构
	TendisTypePredixyTendisplusCluster = "PredixyTendisplusCluster"
	// TendisTypeTwemproxyRedisInstance twemproxy + RedisInstance架构
	TendisTypeTwemproxyRedisInstance = "TwemproxyRedisInstance"
	// TendisTypeTwemproxyTendisplusInstance twemproxy+ TendisplusInstance架构
	TendisTypeTwemproxyTendisplusInstance = "TwemproxyTendisplusInstance"
	// TendisTypeTwemproxyTendisSSDInstance twemproxy+ TendisSSDInstance架构
	TendisTypeTwemproxyTendisSSDInstance = "TwemproxyTendisSSDInstance"
	// TendisTypeRedisInstance RedisCache 主从版
	TendisTypeRedisInstance = "RedisInstance"
	// TendisTypeTendisplusInsance Tendisplus 主从版
	TendisTypeTendisplusInsance = "TendisplusInstance"
	// TendisTypeTendisSSDInsance TendisSSD 主从版
	TendisTypeTendisSSDInsance = "TendisSSDInstance"
	// TendisTypeRedisCluster 原生RedisCluster 架构
	TendisTypeRedisCluster = "RedisCluster"
	// TendisTypeTendisplusCluster TendisplusCluster架构
	TendisTypeTendisplusCluster = "TendisplusCluster"
)

// kibis of bits
const (
	Byte = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
	EiByte
)

const (
	// RedisMasterRole redis role master
	RedisMasterRole = "master"
	// RedisSlaveRole redis role slave
	RedisSlaveRole = "slave"

	// RedisNoneRole none role
	RedisNoneRole = "none"

	// MasterLinkStatusUP up status
	MasterLinkStatusUP = "up"
	// MasterLinkStatusDown down status
	MasterLinkStatusDown = "down"

	// TendisSSDIncrSyncState IncrSync state
	TendisSSDIncrSyncState = "IncrSync"
	// TendisSSDReplFollowtate REPL_FOLLOW  state
	TendisSSDReplFollowtate = "REPL_FOLLOW"
)

const (
	// RedisLinkStateConnected redis connection status connected
	RedisLinkStateConnected = "connected"
	// RedisLinkStateDisconnected redis connection status disconnected
	RedisLinkStateDisconnected = "disconnected"
)

const (
	// NodeStatusPFail Node is in PFAIL state. Not reachable for the node you are contacting, but still logically reachable
	NodeStatusPFail = "fail?"
	// NodeStatusFail Node is in FAIL state. It was not reachable for multiple nodes that promoted the PFAIL state to FAIL
	NodeStatusFail = "fail"
	// NodeStatusHandshake Untrusted node, we are handshaking.
	NodeStatusHandshake = "handshake"
	// NodeStatusNoAddr No address known for this node
	NodeStatusNoAddr = "noaddr"
	// NodeStatusNoFlags no flags at all
	NodeStatusNoFlags = "noflags"
)

const (
	// ClusterStateOK command 'cluster info',cluster_state
	ClusterStateOK = "ok"
	// ClusterStateFail command 'cluster info',cluster_state
	ClusterStateFail = "fail"
)
const (
	// DefaultMinSlots  0
	DefaultMinSlots = 0
	// DefaultMaxSlots 16383
	DefaultMaxSlots = 16383

	// TwemproxyMaxSegment twemproxy max segment
	TwemproxyMaxSegment = 419999
	// TotalSlots 集群总槽数
	TotalSlots = 16384
)

// time layout
const (
	UnixtimeLayout     = "2006-01-02 15:04:05"
	FilenameTimeLayout = "20060102-150405"
	FilenameDayLayout  = "20060102"
	UnixtimeLayoutZone = "2006-01-02T15:04:05-07:00"
)

// account
const (
	MysqlAaccount = "mysql"
	MysqlGroup    = "mysql"
	OSAccount     = "mysql"
	OSGroup       = "mysql"
)

// path dirs
const (
	UsrLocal           = "/usr/local"
	PackageSavePath    = "/data/install"
	Data1Path          = "/data1"
	DataPath           = "/data"
	DbaReportSaveDir   = "/home/mysql/dbareport/"
	RedisReportSaveDir = "/home/mysql/dbareport/redis/"
	ExporterConfDir    = "/home/mysql/.exporter"
	RedisReportLeftDay = 15
)

// tool path
const (
	DbToolsPath              = "/home/mysql/dbtools"
	RedisShakeBin            = "/home/mysql/dbtools/redis-shake"
	RedisSafeDeleteToolBin   = "/home/mysql/dbtools/redisSafeDeleteTool"
	LdbTendisplusBin         = "/home/mysql/dbtools/ldb_tendisplus"
	TredisverifyBin          = "/home/mysql/dbtools/tredisverify"
	TredisBinlogBin          = "/home/mysql/dbtools/tredisbinlog"
	TredisDumpBin            = "/home/mysql/dbtools/tredisdump"
	NetCatBin                = "/home/mysql/dbtools/netcat"
	TendisKeyLifecycleBin    = "/home/mysql/dbtools/tendis-key-lifecycle"
	ZkWatchBin               = "/home/mysql/dbtools/zkwatch"
	ZstdBin                  = "/home/mysql/dbtools/zstd"
	LzopBin                  = "/home/mysql/dbtools/lzop"
	LdbWithV38Bin            = "/home/mysql/dbtools/ldb_with_len.3.8"
	LdbWithV513Bin           = "/home/mysql/dbtools/ldb_with_len.5.13"
	MyRedisCaptureBin        = "/home/mysql/dbtools/myRedisCapture"
	BinlogToolTendisplusBin  = "/home/mysql/dbtools/binlogtool_tendisplus"
	RedisCliBin              = "/home/mysql/dbtools/redis-cli"
	TendisDataCheckBin       = "/home/mysql/dbtools/tendisDataCheck"
	RedisDiffKeysRepairerBin = "/home/mysql/dbtools/redisDiffKeysRepairer"
)

// bk-dbmon path
const (
	BkDbmonPath        = "/home/mysql/bk-dbmon"
	BkDbmonBin         = "/home/mysql/bk-dbmon/bk-dbmon"
	BkDbmonConfFile    = "/home/mysql/bk-dbmon/dbmon-config.yaml"
	BkDbmonPort        = 6677
	BkDbmonHTTPAddress = "127.0.0.1:6677"
)

// backup
const (
	NormalBackupType              = "normal_backup"
	ForeverBackupType             = "forever_backup"
	IBSBackupClient               = "/usr/local/bin/backup_client"
	COSBackupClient               = "/usr/local/backup_client/bin/backup_client"
	COSInfoFile                   = "/home/mysql/.cosinfo.toml"
	BackupTarSplitSize            = "8G"
	RedisFullBackupTAG            = "REDIS_FULL"
	RedisBinlogTAG                = "REDIS_BINLOG"
	RedisForeverBackupTAG         = "DBFILE"
	RedisFullBackupReportType     = "redis_fullbackup"
	RedisBinlogBackupReportType   = "redis_binlogbackup"
	RedisFullbackupRepoter        = "redis_fullbackup_%s.log"
	RedisBinlogRepoter            = "redis_binlog_%s.log"
	BackupStatusStart             = "start"
	BackupStatusRunning           = "running"
	BackupStatusToBakSystemStart  = "to_backup_system_start"
	BackupStatusToBakSystemFailed = "to_backup_system_failed"
	BackupStatusToBakSysSuccess   = "to_backup_system_success"
	BackupStatusFailed            = "failed"
	BackupStatusLocalSuccess      = "local_success"
	BackupClientStrorageTypeCOS   = "cos"
	BackupClientStrorageTypeHDFS  = "hdfs"
)

// BackupSystem
const (
	BackupTaskSuccess        string = "4"
	FileExpired              string = "1"
	FileNotExpired           string = "0"
	BackupVersion            string = "1.0"
	BackupMaxQueryRetryTimes int    = 60
)

// meta role
const (
	MetaRoleRedisMaster = "redis_master"
	MetaRoleRedisSlave  = "redis_slave"
	MetaRolePredixy     = "predixy"
	MetaRoleTwemproxy   = "twemproxy"
)

// proxy operations
const (
	ProxyStart    = "proxy_open"
	ProxyStop     = "proxy_close"
	ProxyRestart  = "proxy_restart"
	ProxyShutdown = "proxy_shutdown"
)

const (
	// FlushDBRename ..
	FlushDBRename = "cleandb"
	// CacheFlushAllRename ..
	CacheFlushAllRename = "cleanall"
	// SSDFlushAllRename ..
	SSDFlushAllRename = "flushalldisk"
	// KeysRename ..
	KeysRename = "mykeys"
	// ConfigRename ..
	ConfigRename = "confxx"
	// TendisPlusFlushAllRename ..
	TendisPlusFlushAllRename = "cleanall"
)

// IsClusterDbType 存储端是否是cluster类型
func IsClusterDbType(dbType string) bool {
	if dbType == TendisTypePredixyRedisCluster ||
		dbType == TendisTypePredixyTendisplusCluster ||
		dbType == TendisTypeRedisCluster ||
		dbType == TendisTypeTendisplusCluster {
		return true
	}
	return false
}

// IsRedisInstanceDbType 存储端是否是cache类型
func IsRedisInstanceDbType(dbType string) bool {
	if dbType == TendisTypePredixyRedisCluster ||
		dbType == TendisTypeTwemproxyRedisInstance ||
		dbType == TendisTypeRedisInstance ||
		dbType == TendisTypeRedisCluster {
		return true
	}
	return false
}

// IsTwemproxyClusterType 检查proxy是否为Twemproxy
func IsTwemproxyClusterType(dbType string) bool {
	if dbType == TendisTypeTwemproxyRedisInstance ||
		dbType == TendisTypeTwemproxyTendisSSDInstance ||
		dbType == TendisTypeTwemproxyTendisplusInstance {
		return true
	}
	return false
}

// IsPredixyClusterType 检查proxy是否为Predixy
func IsPredixyClusterType(dbType string) bool {
	if dbType == TendisTypePredixyRedisCluster ||
		dbType == TendisTypePredixyTendisplusCluster {
		return true
	}
	return false
}

// IsTendisplusInstanceDbType 存储端是否是tendisplus类型
func IsTendisplusInstanceDbType(dbType string) bool {
	if dbType == TendisTypePredixyTendisplusCluster ||
		dbType == TendisTypeTwemproxyTendisplusInstance ||
		dbType == TendisTypeTendisplusInsance ||
		dbType == TendisTypeTendisplusCluster {
		return true
	}
	return false
}

// IsTendisSSDInstanceDbType 存储端是否是tendisSSD类型
func IsTendisSSDInstanceDbType(dbType string) bool {
	if dbType == TendisTypeTwemproxyTendisSSDInstance ||
		dbType == TendisTypeTendisSSDInsance {
		return true
	}
	return false
}

// IsAllowFlushMoreDB 是否支持flush 多DB
func IsAllowFlushMoreDB(dbType string) bool {
	if dbType == TendisTypeRedisInstance ||
		dbType == TendisTypeTendisplusInsance {
		return true
	}
	return false
}

// IsAllowRandomkey 是否支持randomkey命令
func IsAllowRandomkey(dbType string) bool {
	if dbType == TendisTypePredixyTendisplusCluster ||
		dbType == TendisTypeTwemproxyTendisplusInstance ||
		dbType == TendisTypeTendisplusInsance ||
		dbType == TendisTypeTendisplusCluster {
		return false
	}
	return true
}
