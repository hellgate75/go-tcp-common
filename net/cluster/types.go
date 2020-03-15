package cluster

import (
	cio "github.com/hellgate75/go-tcp-common/io"
	"github.com/hellgate75/go-tcp-common/net/cluster/types"
	"net"
	"regexp"
)

type ClusterRegistry interface{
	Register(n *types.Node) error
	Update(field string, filter regexp.Regexp, n types.Node) error
	Recover(field string, filter regexp.Regexp) ([]*types.Node, error)
	List() []types.Node
}

type ClusterNode interface {
	Listen(ip string, port int32) error
	Command(n *types.Node, command types.Command) error
	Aknoledge(n *types.Node) error
	Discover(network *net.IPNet, ports types.Ports)
	Stop() error
	List() []types.Node
	UsedFormat() cio.ParserFormat
}