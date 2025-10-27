package global

import (
	"github.com/gangantongxue/landlady/cmd/model"
	"github.com/gangantongxue/landlady/cmd/opts"
	"github.com/hashicorp/raft"
)

var (
	Landlady *raft.Raft
)

var (
	Opts      = &opts.Opts{}
	NodeNum   = &model.NodeNum{}
	NodesInfo = &model.NodesInfo{}
	FSM       = &model.FSM{}
)
