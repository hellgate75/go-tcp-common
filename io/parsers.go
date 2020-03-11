package io

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

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
