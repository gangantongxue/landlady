package opts

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/gangantongxue/landlady/cmd/version"
)

// Opts 命令行参数
// 集群初始化时所需的参数
type Opts struct {
	// 集群节点数量，默认3个节点，int类型
	// 推荐使用奇数个节点
	// 当节点数量<=NodeNum/2时，集群判定为不可用
	NodeNum int

	// 节点唯一ID，不可重复，string类型
	NodeID string

	// 节点间通信地址，IP:PORT，string类型
	PeerAddr string

	// 客户端请求地址，IP:PORT，string类型
	ClientAddr string

	// 存储路径（文件夹），string类型
	DataDir string

	// 是否加入已有集群，bool类型
	Join bool

	// 引导节点监听端口，string类型
	GuidePort string
	// 加入已有集群时，指定已有集群引导节点地址，IP:PORT，string类型
	GuideAddr string

	// 健康检查间隔，int类型
	HealthCheckInterval int

	// 日志文件最大尺寸(MB)，int类型
	LogMaxSize int

	// 日志文件最大数量，int类型
	LogMaxBackups int

	// 日志文件最大保留天数，int类型
	LogMaxAge int

	// 日志文件是否压缩，bool类型
	LogCompress bool

	// 日志是否输出到控制台，bool类型
	ToConsole bool
}

// Init 初始化参数
// 初始化参数，传入Opts指针，返回error
func Init(o *Opts) error {
	if *flag.Bool("version", false, "version") {
		version.PrintVersion()
		os.Exit(0)
	}

	flag.IntVar(&o.NodeNum, "node-num", 3, "node num")
	flag.StringVar(&o.NodeID, "node-id", "", "node id")
	flag.StringVar(&o.PeerAddr, "peer-addr", "", "peer addr, IP:PORT")
	flag.StringVar(&o.ClientAddr, "client-addr", "", "client addr, IP:PORT")
	flag.StringVar(&o.DataDir, "data-dir", "var/lib/landlady", "data dir")
	flag.BoolVar(&o.Join, "join", false, "join")
	flag.StringVar(&o.GuidePort, "guide-port", "", "guide port")
	flag.StringVar(&o.GuideAddr, "guide-addr", "", "guide addr, IP:PORT")
	flag.IntVar(&o.HealthCheckInterval, "health-check-interval (s)", 5, "health check interval (s)")
	flag.IntVar(&o.LogMaxSize, "log-max-size (MB)", 100, "log max size (MB)")
	flag.IntVar(&o.LogMaxBackups, "log-max-backups", 7, "log max backups")
	flag.IntVar(&o.LogMaxAge, "log-max-age", 30, "log max age")
	flag.BoolVar(&o.LogCompress, "log-compress", true, "log compress")
	flag.BoolVar(&o.ToConsole, "log-to-console", false, "log to console")

	flag.Parse()

	if o.NodeID == "" {
		return errors.New("node id is empty")
	}
	if o.PeerAddr == "" {
		return errors.New("peer addr is empty")
	}
	if o.ClientAddr == "" {
		return errors.New("client addr is empty")
	}
	if o.Join {
		if o.GuideAddr == "" {
			return errors.New("guide addr is empty")
		}
	}
	if o.GuidePort == "" {
		o.GuidePort = strings.Split(o.ClientAddr, ":")[1]
	}
	if o.NodeNum <= 0 {
		return errors.New("node num must be greater than 0")
	}
	return nil
}
