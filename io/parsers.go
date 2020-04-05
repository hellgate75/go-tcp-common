package io

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/common"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"plugin"
	"strings"
)

type ParserFormat string

const(
	ParserFormatJson    ParserFormat = "JSON"
	ParserFormatXml     ParserFormat = "XML"
	ParserFormatYaml    ParserFormat = "YAML"

)

var DefaultPluginsFolder string = common.GetCurrentPath() + string(os.PathListSeparator) + "modules"
var DefaultLibraryExtension string = common.GetShareLibExt()

// Describe the exposed Plugin interface proxy function (expected function name = ParserPlugin
type ParserPlugin func(format ParserFormat) (FormatParser, error)

// Describe the exposed Plugin interface proxy function (expected function name = PluginsCollector
type PluginsCollector func() ([]FormatParser, error)

//Describes as a Plugin Parser executive must expose functions
type FormatParser interface{
	// Provided format
	ParserFormat() ParserFormat
	// Mashal an interface to the specified format in bytes
	Marshall(itf interface{}) ([]byte, error)
	// Mashal a byte array into the specified interface, returning the unmashalled element
	Unmashall(code []byte, itf interface{}) (interface{}, error)
}

var pluginsCache map[ParserFormat]FormatParser

var pluginsList []FormatParser = make([]FormatParser, 0)

//Looks up for plugin by a given ParserFormat, using  eventually pluginsFolder and library exception excluded the dot,
// in case one of them is empty string it will be replaced with the package DefaultPluginsFolder and DefaultLibraryExtension variables
func CollectAllPlugins(pluginsFolder string, libExtension string) ([]FormatParser, error) {
	if len(pluginsList) > 0 {
		return pluginsList, nil
	}
	var out []FormatParser = make([]FormatParser, 0)
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("io.CollectAllPlugins - Unable to read plugins, Details: %v", r))
		}
	}()
	if "" == pluginsFolder {
		pluginsFolder = DefaultPluginsFolder
	}
	if "" == libExtension {
		pluginsFolder = DefaultLibraryExtension
	}
	if ! ExistsFile(pluginsFolder) {
		return out, errors.New(fmt.Sprintf("io.CollectAllPlugins - File %s doesn't exist!!", pluginsFolder))
	}
	lext := strings.ToLower(fmt.Sprintf(".%s", libExtension))
	if IsFolder(pluginsFolder) {
		libraries := GetMatchedFiles(pluginsFolder, true, func(path string) bool{
			return strings.ToLower(path[len(path)-len(lext):]) == lext
		})
		for _, lib := range libraries {
			plugin, err := plugin.Open(lib)
			if err == nil {
				symbol, err := plugin.Lookup("PluginsCollector")
				if err == nil {
					parserPlugin := symbol.(PluginsCollector)
					parsers, err := parserPlugin()
					if err == nil {
						for _, parser := range parsers {
							if parser == nil {
								out = append(out, parser)
							}
						}
					}
				}

			}
		}
	} else {
		return out, errors.New(fmt.Sprintf("io.CollectAllPlugins - File %s is not a folder!!", pluginsFolder))
	}
	return out, err
}

//Looks up for plugin by a given ParserFormat, using  eventually pluginsFolder and library exception excluded the dot,
// in case one of them is empty string it will be replaced with the package DefaultPluginsFolder and DefaultLibraryExtension variables
func LookupInPlugins(format ParserFormat, pluginsFolder string, libExtension string) (FormatParser, error) {
	if plugin, ok := pluginsCache[format]; ok {
		return plugin, nil
	}
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("io.LookupInPlugins - Unable to read plugins, Details: %v", r))
		}
	}()
	if "" == pluginsFolder {
		pluginsFolder = DefaultPluginsFolder
	}
	if "" == libExtension {
		pluginsFolder = DefaultLibraryExtension
	}
	if ! ExistsFile(pluginsFolder) {
		return nil, errors.New(fmt.Sprintf("io.LookupInPlugins - File %s doesn't exist!!", pluginsFolder))
	}
	lext := strings.ToLower(fmt.Sprintf(".%s", libExtension))
	if IsFolder(pluginsFolder) {
		libraries := GetMatchedFiles(pluginsFolder, true, func(path string) bool{
			return strings.ToLower(path[len(path)-len(lext):]) == lext
		})
		for _, lib := range libraries {
			plugin, err := plugin.Open(lib)
			if err == nil {
				symbol, err := plugin.Lookup("ParserPlugin")
				if err == nil {
					parserPlugin := symbol.(ParserPlugin)
					parser, err := parserPlugin(format)
					if err == nil {
						pluginsCache[format]=parser
						return parser, nil
					}
				}

			}
		}
	} else {
		return nil, errors.New(fmt.Sprintf("io.LookupInPlugins - File %s is not a folder!!", pluginsFolder))
	}
	if err != nil {
		return nil ,err
	}
	return nil, errors.New(fmt.Sprintf("io.LookupInPlugins - No match found for parser format: %v", format))
}

// Marshall an object instance tranforming in byte array, reporting eventually errors based on
// the required parser format
func Marshall(itf interface{}, format ParserFormat) ([]byte, error) {
	var text string = ""
	var err error = nil
	if strings.ToUpper(string(format)) == string(ParserFormatJson) {
		text, err = ToJson(itf)
	} else if strings.ToUpper(string(format)) == string(ParserFormatYaml) {
		text, err = ToYaml(itf)
	} else if strings.ToUpper(string(format)) == string(ParserFormatXml) {
		text, err = ToXml(itf)
	} else if parser, err := LookupInPlugins(format, "", ""); err == nil && parser != nil {
		bytes, errB := parser.Marshall(itf)
		if errB != nil {
			err = errB
			text = ""
		} else {
			err = nil
			text = string(bytes)
		}
	} else {
		return []byte{}, errors.New(fmt.Sprintf("Unable to identify following parser format: %v", format))
	}
	return []byte(text), err
}

// Marshall an object instance tranforming in byte array and saving in a file in an existing path,
// reporting eventually errors based on the required parser format
func MarshallTo(itf interface{}, filePath string, format ParserFormat) error {
	text, err := Marshall(itf, format)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to identify following parser format: %v", format))
	}
	if err == nil {
		ioutil.WriteFile(filePath, []byte(text), 0660)
	}
	return err
}

// Marshall data byte arrays parsing the data into the interface returned of same type of the
// given one, reporting eventually errors based on the required parser format
func Unmashall(code []byte, itf interface{}, format ParserFormat) (interface{}, error) {
	var err error = nil
	if strings.ToUpper(string(format)) == string(ParserFormatJson) {
		itf, err = FromJsonCode(string(code), itf)
	} else if strings.ToUpper(string(format)) == string(ParserFormatYaml) {
		itf, err = FromYamlCode(string(code), itf)
	} else if strings.ToUpper(string(format)) == string(ParserFormatXml) {
		itf, err = FromXmlCode(string(code), itf)
	} else if parser, err := LookupInPlugins(format, "", ""); err ==nil && parser != nil {
		itf, err = parser.Unmashall(code, itf)
	} else {
		return nil, errors.New(fmt.Sprintf("Unable to identify following parser format: %v", format))
	}
	return itf, err

}


// Marshall file byte arrays parsing the data into the interface returned of same type of the
// given one, reporting eventually errors based on the required parser format
func UnmashallFrom(filePath string, itf interface{}, format ParserFormat) (interface{}, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	file, errF := os.Open(filePath)
	if errF != nil {
		return nil, errF
	}
	data, errR := ioutil.ReadAll(file)
	if errR != nil {
		return nil, errR
	}
	return Unmashall(data, itf, format)
	
}

// Trasform Yaml code in Object
func FromYamlCode(yamlCode string, itf interface{}) (interface{}, error) {
	err := yaml.Unmarshal([]byte(yamlCode), itf)
	if err != nil {
		return nil, errors.New("FromYamlCode::Unmarshal: " + err.Error())
	} else {
		return itf, nil
	}
}

// Trasform Json code in Object
func FromJsonCode(jsonCode string, itf interface{}) (interface{}, error) {
	err := json.Unmarshal([]byte(jsonCode), itf)
	if err != nil {
		return nil, errors.New("FromJsonCode::Unmarshal: " + err.Error())
	} else {
		return itf, nil
	}
}

// Trasform Xml code in Object
func FromXmlCode(xmlCode string, itf interface{}) (interface{}, error) {
	err := xml.Unmarshal([]byte(xmlCode), itf)
	if err != nil {
		return nil, errors.New("FromXmlCode::Unmarshal: " + err.Error())
	} else {
		return itf, nil
	}
}

// Load a Yaml file and transform it in Object
func FromYamlFile(path string, itf interface{}) (interface{}, error) {
	_, errS := os.Stat(path)
	if errS != nil {
		return nil, errors.New("FromYamlFile::Stats: " + errS.Error())
	}
	file, errF := os.Open(path)
	if errF != nil {
		return nil, errors.New("FromYamlFile::OpenFile: " + errF.Error())
	}
	bytes, errR := ioutil.ReadAll(file)
	if errR != nil {
		return nil, errors.New("FromYamlFile::ReadFile: " + errR.Error())
	}
	err := yaml.Unmarshal(bytes, itf)
	if err != nil {
		return nil, errors.New("FromYamlFile::Unmarshal: " + err.Error())
	} else {
		return itf, nil
	}
}

// Load a JSON file and transform it in Object
func FromJsonFile(path string, itf interface{}) (interface{}, error) {
	_, errS := os.Stat(path)
	if errS != nil {
		return nil, errors.New("FromJsonFile::Stats: " + errS.Error())
	}
	file, errF := os.Open(path)
	if errF != nil {
		return nil, errors.New("FromJsonFile::OpenFile: " + errF.Error())
	}
	bytes, errR := ioutil.ReadAll(file)
	if errR != nil {
		return nil, errors.New("FromJsonFile::ReadFile: " + errR.Error())
	}
	err := json.Unmarshal(bytes, itf)
	if err != nil {
		return nil, errors.New("FromJsonFile::Unmarshal: " + err.Error())
	} else {
		return itf, nil
	}
}

// Load an Xml file and transform it in Object
func FromXmlFile(path string, itf interface{}) (interface{}, error) {
	_, errS := os.Stat(path)
	if errS != nil {
		return nil, errors.New("FromXmlFile::Stats: " + errS.Error())
	}
	file, errF := os.Open(path)
	if errF != nil {
		return nil, errors.New("FromXmlFile::OpenFile: " + errF.Error())
	}
	bytes, errR := ioutil.ReadAll(file)
	if errR != nil {
		return nil, errors.New("FromXmlFile::ReadFile: " + errR.Error())
	}
	err := xml.Unmarshal(bytes, itf)
	if err != nil {
		return nil, errors.New("FromXmlFile::Unmarshal: " + err.Error())
	} else {
		return itf, nil
	}
}

// Transform an interface in Yaml Code
func ToYaml(itf interface{}) (string, error) {
	bytes, err := yaml.Marshal(itf)
	if err != nil {
		return "", errors.New("ToYaml::Marshal: " + err.Error())
	} else {
		return fmt.Sprintf("\n%s", bytes), nil
	}
}

// Transform an interface in JSON Code
func ToJson(itf interface{}) (string, error) {
	bytes, err := json.Marshal(itf)
	if err != nil {
		return "", errors.New("ToJson::Marshal: " + err.Error())
	} else {
		return fmt.Sprintf("\n%s", bytes), nil
	}
}

// Transform an interface in XML Code
func ToXml(itf interface{}) (string, error) {
	bytes, err := xml.MarshalIndent(itf, "", "  ")
	if err != nil {
		return "", errors.New("ToXml::Marshal: " + err.Error())
	} else {
		return fmt.Sprintf("\n%s", bytes), nil
	}
}
