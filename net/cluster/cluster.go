package cluster

import "github.com/hellgate75/go-tcp-common/net/api/common"

type clusterNode struct {
	PluginsFolder		string				`yaml:"pluginsFolder,omitempty" json:"pluginsFolder,omitempty" xml:"plugins-folder,omitempty"`
	PluginsExtension	string				`yaml:"pluginsExtension,omitempty" json:"pluginsExtension,omitempty" xml:"plugins-extension,omitempty"`
	PluginsEnabled		bool				`yaml:"pluginsEnabled,omitempty" json:"pluginsEnabled,omitempty" xml:"plugins-enabled,omitempty"`
	_apiServer			common.ApiServer	`yaml:"-,omitempty" json:"-,omitempty" xml:"-,omitempty"`
	
}

