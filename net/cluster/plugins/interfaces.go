package plugins

import (
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/io"
	"github.com/hellgate75/go-tcp-common/net/cluster/types"
	"github.com/hellgate75/go-tcp-common/net/common"
	comm "github.com/hellgate75/go-tcp-common/common"
	"os"
	"plugin"
	"strings"
)

// Describe the exposed Plugin interface for the proxies access to the service triplets ( made of: Path string, net/common.ApiAction, net/cluster/types.Command)
type ServicePlugin interface {
	GetActionAndPath() (string, common.ApiAction, *types.Command, error)
}

// Describe Proxy Lookup function, expected with name = ServiceDiscovery -> func()([]ServicePlugin)
type ServiceDiscovery func()([]ServicePlugin, error)


var DefaultPluginsFolder string = comm.GetCurrentPath() + string(os.PathListSeparator) + "services"
var DefaultLibraryExtension string = comm.GetShareLibExt()

var pluginsList []ServicePlugin = make([]ServicePlugin, 0)

//Looks up for plugin by a given ParserFormat, using  eventually pluginsFolder and library exception excluded the dot,
// in case one of them is empty string it will be replaced with the package DefaultPluginsFolder and DefaultLibraryExtension variables
func CollectAllPlugins(pluginsFolder string, libExtension string) ([]ServicePlugin, error) {
	if len(pluginsList) > 0 {
		return pluginsList, nil
	}
	var out []ServicePlugin = make([]ServicePlugin, 0)
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("net/cluster/plugins.CollectAllPlugins - Unable to connect given nodes, Details: %v", r))
		}
	}()
	if "" == pluginsFolder {
		pluginsFolder = DefaultPluginsFolder
	}
	if "" == libExtension {
		pluginsFolder = DefaultLibraryExtension
	}
	if ! io.ExistsFile(pluginsFolder) {
		return out, errors.New(fmt.Sprintf("net/cluster/plugins.CollectAllPlugins - File %s doesn't exist!!", pluginsFolder))
	}
	lext := strings.ToLower(fmt.Sprintf(".%s", libExtension))
	if io.IsFolder(pluginsFolder) {
		libraries := io.GetMatchedFiles(pluginsFolder, true, func(path string) bool{
			return strings.ToLower(path[len(path)-len(lext):]) == lext
		})
		for _, lib := range libraries {
			plugin, err := plugin.Open(lib)
			if err == nil {
				symbol, err := plugin.Lookup("ServiceDiscovery")
				if err == nil {
					parserPlugin := symbol.(ServiceDiscovery)
					plugins, err := parserPlugin()
					if err == nil {
						for _, plugin := range plugins {
							if plugin == nil {
								out = append(out, plugin)
							}
						}
					}
				}

			}
		}
	} else {
		return out, errors.New(fmt.Sprintf("net/cluster/plugins.CollectAllPlugins - File %s is not a folder!!", pluginsFolder))
	}
	return out, err
}
