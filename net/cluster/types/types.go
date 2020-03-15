package types

import (
	"github.com/hellgate75/go-tcp-common/net/common"
	"time"
)

type NodeType byte
type PortType byte
type NodeState byte

const (
	ROLE_UNKNOWN NodeType = 0
	ROLE_MASTER  NodeType = iota + 1
	ROLE_SLAVE
	ROLE_CCORDINATOR
	PORT_TYPE_UNKNOWN PortType = 0
	PORT_TYPE_REST    PortType = iota + 1
	PORT_TYPE_STREAM
	PORT_TYPE_ROW
	NODE_STATE_UNKNOWN NodeState = 0
	NODE_STATE_RUNNING NodeState = iota + 1
	NODE_STATE_PAUSED
	NODE_STATE_UNRACJABLE
)

type Ports struct {
	MinPort int32
	MaxPort int32
}

type Command struct {
	Name      string
	Command   string
	Arguments []string
	Method    common.RestMethod
	Accepts   common.MimeType
	Produces  common.MimeType
}

type Service struct {
	Port     Port
	Commands []Command
}

type Port struct {
	Port  int32
	Descr string
	Type  PortType
}

type Node struct {
	Name      string
	IpAddress string
	Ports     []Port
	Role      NodeType
	Services  []Service
	Active    bool
	LastCheck time.Time
	State     NodeState
	Info      NodeInfo
}

type NodeInfo struct {
	OS       string `yaml:"osName,omitempty" json:"osName,omitempty" xml:"os-name,chardata,omitempty"`
	Arch     string `yaml:"osArch,omitempty" json:"osArch,omitempty" xml:"os-arch,chardata,omitempty"`
	GoPath   string `yaml:"goPath,omitempty" json:"goPath,omitempty" xml:"go-path,chardata,omitempty"`
	NumCPUs  int    `yaml:"numCpus,omitempty" json:"numCpus,omitempty" xml:"num-cpus,chardata,omitempty"`
	Timezone string `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,chardata,omitempty"`
}

