package discovery

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/io"
	"github.com/hellgate75/go-tcp-common/net/cluster/types"
	"github.com/hellgate75/go-tcp-common/net/common"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func DiscoverNodes(network *net.IPNet, timeout time.Duration, netType string, ports types.Ports, tlsConfig *tls.Config, collectInfo func()()) ([]types.NodePingInfo, error) {
	var out []types.NodePingInfo = make([]types.NodePingInfo, 0)
	if "" == netType {
		netType="tcp"
	}
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("net/cluster/discover.DiscoverNodes - Unable to discover nodes, Details: %v", r))
		}
	}()
	addressList := common.ListAddresses(network)
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	proto := "http"
	if tlsConfig != nil {
		proto = "https"
	}
	defer func(){
		time.Sleep(5 * time.Second)
		client.CloseIdleConnections()
	}()
	for _, address := range addressList {
		ip := address.String()
		for port := ports.MinPort; port<=ports.MaxPort; port++ {
			addressPort := fmt.Sprintf("%s:%v", ip, port)
			url := fmt.Sprintf("%s://%s/ping", proto, addressPort)
			var nodePingInfo *types.NodePingInfo
			response, err := client.Get(url)
			if err == nil {
				data, err := ioutil.ReadAll(response.Body)
				if err == nil {
					nodePingInfo = parseNodePingInfoWithAllFormats(data)
					if nodePingInfo != nil {
						nodePingInfo.IpAddress = ip
						nodePingInfo.Port = port
						nodePingInfo.Time = time.Now()
						out = append(out, *nodePingInfo)
					}
				}
			}
		}

	}
	return out, err
}

func RequireServiceInfo(nodesInfoList []types.NodePingInfo, timeout time.Duration, tlsConfig *tls.Config) ([]types.Node, error) {
	var err error
	var out []types.Node = make([]types.Node, 0)
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("net/cluster/discover.RequireServiceInfo - Unable to connect given nodes, Details: %v", r))
		}
	}()
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	defer func(){
		time.Sleep(5 * time.Second)
		client.CloseIdleConnections()
	}()
	proto := "http"
	if tlsConfig != nil {
		proto = "https"
	}
	for _, nodePingInfo := range nodesInfoList {
		service := fmt.Sprintf("%s:%v", nodePingInfo.IpAddress, nodePingInfo.Port)
		url := fmt.Sprintf("%s://%s/info", proto, service)
		response, err := client.Get(url)
		if err == nil {
			var nodeInfo *types.NodeInfo
			if ( response.StatusCode == 200) {
				data, err := ioutil.ReadAll(response.Body)
				if err == nil {
					nodeInfo = parseNodeInfoWithAllFormats(data)
				}

			}
			var services []types.Service = make([]types.Service, 0)
			url2 := fmt.Sprintf("%s://%s/services", proto, service)
			response2, err := client.Get(url2)
			if err == nil {
				if response2.StatusCode == 200 {
					data, err := ioutil.ReadAll(response2.Body)
					if err == nil {
						services = append(services, parseServicesWithAllFormats(data)...)
					}
				}
			}
			node := types.Node{
				Name: fmt.Sprintf("%s_%v", nodePingInfo.IpAddress, nodePingInfo.Port),
				IpAddress: nodePingInfo.IpAddress,
				Port: nodePingInfo.Port,
				Ports: nodePingInfo.Ports,
				Role: nodePingInfo.Role,
				Services: services,
				State: nodePingInfo.State,
				Active: nodePingInfo.Active,
				LastCheck: nodePingInfo.Time,
			}
			if nodeInfo != nil {
				node.Info = *nodeInfo
			}
			out = append(out, node)
		}
	}
	return out, err
}

var parsersCache []io.FormatParser = make([]io.FormatParser, 0)

func parseNodePingInfoWithAllFormats(code []byte) *types.NodePingInfo {
	var itfIn = types.NodePingInfo{}
	itf, err := io.FromJsonCode(string(code), itfIn)
	if err == nil {
		var nodePingInfo types.NodePingInfo = itf.(types.NodePingInfo)
		return &nodePingInfo
	}
	itf, err = io.FromYamlCode(string(code), itfIn)
	if err == nil {
		var nodePingInfo types.NodePingInfo = itf.(types.NodePingInfo)
		return &nodePingInfo
	}
	itf, err = io.FromXmlCode(string(code), itfIn)
	if err == nil {
		var nodePingInfo types.NodePingInfo = itf.(types.NodePingInfo)
		return &nodePingInfo
	}
	var parsers []io.FormatParser = make([]io.FormatParser, 0)
	if len(parsersCache) > 0 {
		parsers = parsersCache
	} else {
		parsers, err = io.CollectAllPlugins("", "")
		if err != nil {
			return nil
		}
		parsersCache = append(parsersCache, parsers...)
	}
	for _, parser := range parsers {
		itf, err = parser.Unmashall(code, itfIn)
		if err == nil {
			var nodePingInfo types.NodePingInfo = itf.(types.NodePingInfo)
			return &nodePingInfo
		}
	}
	return nil
}

func parseServicesWithAllFormats(code []byte) []types.Service {
	var itfIn = make([]types.Service, 0)
	itf, err := io.FromJsonCode(string(code), itfIn)
	if err == nil {
		var services []types.Service = itf.([]types.Service)
		return services
	}
	itf, err = io.FromYamlCode(string(code), itfIn)
	if err == nil {
		var services []types.Service = itf.([]types.Service)
		return services
	}
	itf, err = io.FromXmlCode(string(code), itfIn)
	if err == nil {
		var services []types.Service = itf.([]types.Service)
		return services
	}
	var parsers []io.FormatParser = make([]io.FormatParser, 0)
	if len(parsersCache) > 0 {
		parsers = parsersCache
	} else {
		parsers, err = io.CollectAllPlugins("", "")
		if err != nil {
			return nil
		}
		parsersCache = append(parsersCache, parsers...)
	}
	for _, parser := range parsers {
		itf, err = parser.Unmashall(code, itfIn)
		if err == nil {
			var services []types.Service = itf.([]types.Service)
			return services
		}
	}
	return itfIn
}

func parseNodeInfoWithAllFormats(code []byte) *types.NodeInfo {
	var itfIn = types.NodeInfo{}
	itf, err := io.FromJsonCode(string(code), itfIn)
	if err == nil {
		var nodeInfo types.NodeInfo = itf.(types.NodeInfo)
		return &nodeInfo
	}
	itf, err = io.FromYamlCode(string(code), itfIn)
	if err == nil {
		var nodeInfo types.NodeInfo = itf.(types.NodeInfo)
		return &nodeInfo
	}
	itf, err = io.FromXmlCode(string(code), itfIn)
	if err == nil {
		var nodeInfo types.NodeInfo = itf.(types.NodeInfo)
		return &nodeInfo
	}
	var parsers []io.FormatParser = make([]io.FormatParser, 0)
	if len(parsersCache) > 0 {
		parsers = parsersCache
	} else {
		parsers, err = io.CollectAllPlugins("", "")
		if err != nil {
			return nil
		}
		parsersCache = append(parsersCache, parsers...)
	}
	for _, parser := range parsers {
		itf, err = parser.Unmashall(code, itfIn)
		if err == nil {
			var nodeInfo types.NodeInfo = itf.(types.NodeInfo)
			return &nodeInfo
		}
	}
	return nil
}