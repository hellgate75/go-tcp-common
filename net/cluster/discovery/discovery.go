package discovery

import (
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/net/cluster/types"
	"net"
)

func DiscoverServers(network *net.IPNet, ports types.Ports, collectInfo func()()) ([]types.Node, error) {
	var nodes []types.Node = make([]types.Node, 0)
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("Unable to discover nodes, Details: %v", r))
		}
	}()

	return nodes, err
}
