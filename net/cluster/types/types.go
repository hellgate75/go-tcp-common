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
	Port      int32
	Ports     []Port
	Role      NodeType
	Services  []Service
	Active    bool
	LastCheck time.Time
	State     NodeState
	Info      NodeInfo
}

type NodeInfo struct {
	OS       string    `yaml:"osName,omitempty" json:"osName,omitempty" xml:"os-name,chardata,omitempty"`
	Arch     string    `yaml:"osArch,omitempty" json:"osArch,omitempty" xml:"os-arch,chardata,omitempty"`
	GoPath   string    `yaml:"goPath,omitempty" json:"goPath,omitempty" xml:"go-path,chardata,omitempty"`
	NumCPUs  int       `yaml:"numCpus,omitempty" json:"numCpus,omitempty" xml:"num-cpus,chardata,omitempty"`
	Timezone string    `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,chardata,omitempty"`
}

type NodePingInfo struct {
	Role      NodeType  `yaml:"role,omitempty" json:"role,omitempty" xml:"role,chardata,omitempty"`
	State     NodeState `yaml:"state,omitempty" json:"state,omitempty" xml:"state,chardata,omitempty"`
	Active	  bool 		`yaml:"active,omitempty" json:"active,omitempty" xml:"active,chardata,omitempty"`
	Ports     []Port	`yaml:"ports,omitempty" json:"ports,omitempty" xml:"ports,chardata,omitempty"`
	IpAddress string    `yaml:"-" json:"-" xml:"-"`
	Port	  int32     `yaml:"-" json:"-" xml:"-"`
	Time	  time.Time `yaml:"-" json:"-" xml:"-"`
}
