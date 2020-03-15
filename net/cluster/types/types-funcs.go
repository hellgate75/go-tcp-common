package types

import "fmt"

func (ni NodeInfo) String() string {
	return fmt.Sprintf("NodeInfo{OS: \"%s\", Arch: \"%s\", GoPath: \"%s\", NumCPUs: %v, TimeZone: \"%s\"}",
		ni.OS, ni.Arch, ni.GoPath, ni.NumCPUs, ni.Timezone)
}

func (nt NodeType) String() string {
	switch nt {
	case ROLE_MASTER:
		return "Master"
	case ROLE_SLAVE:
		return "Slave"
	case ROLE_CCORDINATOR:
		return "Coordinator"
	default:
		return "Unknown"
	}
}

func (pt PortType) String() string {
	switch pt {
	case PORT_TYPE_REST:
		return "Rest Port"
	case PORT_TYPE_STREAM:
		return "Data Stream Port"
	case PORT_TYPE_ROW:
		return "Row Port"
	default:
		return "Unknown"
	}
}

func (ns NodeState) String() string {
	switch ns {
	case NODE_STATE_RUNNING:
		return "Running"
	case NODE_STATE_PAUSED:
		return "Maintainance"
	case NODE_STATE_UNRACJABLE:
		return "Unreachable"
	default:
		return "Unknown"
	}
}

func (n *Node) Update(n1 *Node) {
	if n1 == nil {
		return
	}
	if "" != n1.Name {
		n.Name = n1.Name
	}
	if "" != n1.Name {
		n.IpAddress = n1.IpAddress
	}
	if nil != n1.Ports {
		n.Ports = n1.Ports
	}
	if 0 != n1.Role {
		n.Role = n1.Role
	}
	if nil != n1.Services {
		n.Services = n1.Services
	}
	n.Active = n1.Active
	if 0 > n1.LastCheck.Unix() {
		n.LastCheck = n1.LastCheck
	}
	if 0 != n1.State {
		n.State = n1.State
	}
}

