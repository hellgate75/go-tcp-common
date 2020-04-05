package types

import (
	"github.com/hellgate75/go-tcp-common/net/common"
	"os"
	"time"
)

type NodeType byte
type PortType byte
type NodeState byte

var(
	DefaultFilePerm os.FileMode = 0664
	DefaultFolderPerm os.FileMode = 0664
)

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
	Name      string 				`yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	Command   string				`yaml:"command,omitempty" json:"command,omitempty" xml:"command,omitempty"`
	Arguments []string				`yaml:"arguments,omitempty" json:"arguments,omitempty" xml:"argument,omitempty"`
	Method    []common.RestMethod	`yaml:"methods,omitempty" json:"methods,omitempty" xml:"methods,omitempty"`
	Accepts   common.MimeType		`yaml:"accepts,omitempty" json:"accepts,omitempty" xml:"accepts,omitempty"`
	Produces  common.MimeType		`yaml:"produces,omitempty" json:"produces,omitempty" xml:"produces,omitempty"`
}

type Service struct {
	Port     Port				`yaml:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
	Commands []Command			`yaml:"commands,omitempty" json:"commands,omitempty" xml:"commandGroup,omitempty"`
}

type Port struct {
	Port  int32					`yaml:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
	Description string			`yaml:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	Type  PortType				`yaml:"type,omitempty" json:"type,omitempty" xml:"type,omitempty"`
}

type Node struct {
	Name      string 			`yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	IpAddress string 			`yaml:"ipAddress,omitempty" json:"ipAddress,omitempty" xml:"ip-address,omitempty"`
	Port      int32				`yaml:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
	Ports     []Port			`yaml:"ports,omitempty" json:"ports,omitempty" xml:"portGroup,omitempty"`
	Role      NodeType			`yaml:"ports,omitempty" json:"port,omitempty" xml:"ports,omitempty"`
	Services  []Service			`yaml:"services,omitempty" json:"services,omitempty" xml:"serviceGroup,omitempty"`
	Active    bool				`yaml:"active,omitempty" json:"active,omitempty" xml:"active,omitempty"`
	LastCheck time.Time			`yaml:"lastCheck,omitempty" json:"lastCheck,omitempty" xml:"last-check,omitempty"`
	State     NodeState			`yaml:"state,omitempty" json:"state,omitempty" xml:"state,omitempty"`
	Info      NodeInfo			`yaml:"info,omitempty" json:"info,omitempty" xml:"infoGroup,omitempty"`
}

type NodeInfo struct {
	OS       string    			`yaml:"osName,omitempty" json:"osName,omitempty" xml:"os-name,omitempty"`
	Arch     string    			`yaml:"osArch,omitempty" json:"osArch,omitempty" xml:"os-arch,omitempty"`
	GoPath   string    			`yaml:"goPath,omitempty" json:"goPath,omitempty" xml:"go-path,omitempty"`
	NumCPUs  int       			`yaml:"numCpus,omitempty" json:"numCpus,omitempty" xml:"num-cpus,omitempty"`
	Timezone string    			`yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty"`
}

type NodePingInfo struct {
	Role      NodeType  		`yaml:"role,omitempty" json:"role,omitempty" xml:"role,omitempty"`
	State     NodeState 		`yaml:"state,omitempty" json:"state,omitempty" xml:"state,omitempty"`
	Active	  bool 				`yaml:"active,omitempty" json:"active,omitempty" xml:"active,omitempty"`
	Ports     []Port			`yaml:"ports,omitempty" json:"ports,omitempty" xml:"ports,omitempty"`
	IpAddress string    		`yaml:"-" json:"-" xml:"-"`
	Port	  int32     		`yaml:"-" json:"-" xml:"-"`
	Time	  time.Time 		`yaml:"-" json:"-" xml:"-"`
	Answer	  time.Duration 	`yaml:"-" json:"-" xml:"-"`
}

type NodeRequest struct {
	IpAddress		string
	Ports			[]int32
}
