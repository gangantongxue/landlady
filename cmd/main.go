package main

import (
	"github.com/gangantongxue/ggl"
	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/lraft"
	"github.com/gangantongxue/landlady/cmd/opts"
	"path/filepath"
)

func main() {
	cfg := &ggl.Config{
		LogFileName:   "landlady_log_2006-01-02.log",
		LogFileDir:    filepath.Join(global.Opts.DataDir, "log"),
		LogMaxSize:    global.Opts.LogMaxSize,
		LogMaxBackups: global.Opts.LogMaxBackups,
		LogMaxAge:     global.Opts.LogMaxAge,
		LogCompress:   global.Opts.LogCompress,
		ToConsole:     global.Opts.ToConsole,
	}
	d := ggl.New(cfg)
	defer d.Stop()

	if err := opts.Init(global.Opts); err != nil {
		ggl.Panic("opts init failed", ggl.Err(err))
	}
	if err := lraft.Init(); err != nil {
		ggl.Panic("raft init failed", ggl.Err(err))
	}

	lraft.InitHealthCheck()
}
