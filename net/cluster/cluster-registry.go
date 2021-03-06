package cluster

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	cio "github.com/hellgate75/go-tcp-common/io"
	"github.com/hellgate75/go-tcp-common/net/cluster/types"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

type nodeCache struct {
	sync.Mutex
	FilePath string				`yaml:"filePath,omitempty" json:"filePath,omitempty" xml:"file-path,omitempty"`
	Nodes    []types.Node		`yaml:"nodes,omitempty" json:"nodes,omitempty" xml:"node,omitempty"`
	Encoding cio.ParserFormat	`yaml:"encoding,omitempty" json:"encoding,omitempty" xml:"encoding,omitempty"`
}

func (nc *nodeCache) RegistryFilePath() string {
	return nc.FilePath
}

func (nc *nodeCache) RegistryFileEncodingFormat() cio.ParserFormat {
	return nc.Encoding
}

func (nc *nodeCache) ChangeEncodingFormat(encodingFormat cio.ParserFormat) error {
	if "" == string(encodingFormat) {
		return errors.New("DiscoverReporter.ChangeEncodingFormat - Invalid empty encoding format!!")
	}
	var err error = nil
	var locked bool = false
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.ChangeEncodingFormat - Error: %v", r))
		}
		if locked {
			nc.Unlock()
		}
	}()
	nc.Encoding = encodingFormat
	if nc.IsPersistenceEnabled() {
		nc.Lock()
		locked = true
		err = nc.save()
		nc.Unlock()
		locked = false
	}
	return err
}


func (nc *nodeCache) EnablePersistence(registryFile string) error {
	if nc.IsPersistenceEnabled() {
		return errors.New(fmt.Sprintf("DiscoverReporter.EnablePersistence - Persistence already enabled on: %s", nc.FilePath))
	}
	if "" == registryFile {
		return errors.New("DiscoverReporter.EnablePersistence - Invalid empty file for registry persistence!!")
	}
	var err error
	var locked bool = false
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.EnablePersistence - Error: %v", r))
		}
		if locked {
			nc.Unlock()
		}
	}()
	nc.Lock()
	locked = true
	if "" == nc.Encoding {
		nc.Encoding = cio.ParserFormatYaml
	}
	nc.FilePath = registryFile
	err = nc.save()
	nc.Unlock()
	locked = false
	return err
}

func (nc *nodeCache) DisablePersistence() error {
	if ! nc.IsPersistenceEnabled() {
		return errors.New("DiscoverReporter.DisablePersistence - Persistence is not enabled!!")
	}
	var err error
	var locked bool = false
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.EnablePersistence - Error: %v", r))
		}
		if locked {
			nc.Unlock()
		}
	}()
	nc.Lock()
	locked = true
	os.Remove(nc.FilePath)
	nc.FilePath = ""
	nc.Unlock()
	locked = false
	return err
}

func (nc *nodeCache) IsPersistenceEnabled() bool {
	return "" != nc.FilePath
}

func (nc *nodeCache) Register(n *types.Node) error {
	if n == nil {
		return errors.New("DiscoverReporter.Register - Nil node reference")
	}
	var err error
	var locked bool = false
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.Register - Error: %v", r))
		}
		if locked {
			nc.Unlock()
		}
	}()
	nc.Nodes = append(nc.Nodes, *n)
	if nc.IsPersistenceEnabled() {
		nc.Lock()
		locked = true
		err = nc.save()
		nc.Unlock()
		locked = false
	}
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.Register - Error: %s", err))
	}
	return err
}
func (nc *nodeCache) Update(field string, filter regexp.Regexp, n types.Node) error {
	var err error
	var locked bool = false
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.Update - Error: %v", r))
		}
		if locked {
			nc.Unlock()
		}
	}()
	nodes, err := nc.Recover(field, filter)
	if err != nil {
		return nil
	}
	if len(nodes) > 0 {
		for _, node := range nodes {
			node.Update(&n)
		}
		if nc.IsPersistenceEnabled() {
			nc.Lock()
			locked = true
			err = nc.save()
			nc.Unlock()
			locked = false
		}
	}
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.Update - Error: %s", err))
	}
	return err
}
func (nc *nodeCache) Recover(field string, filter regexp.Regexp) ([]*types.Node, error) {
	var err error
	var out []*types.Node = make([]*types.Node, 0)
	var locked bool = false
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.Recover - Error: %v", r))
		}
		if locked {
			nc.Unlock()
		}
	}()
	if nc.IsPersistenceEnabled() {
		nc.Lock()
		locked  = true
		err = nc.load()
		nc.Unlock()
		locked  = false
		if err != nil {
			return out, err
		}
	}
	if strings.Index(field, ".") > 0{
		if len(field) > 9 && strings.ToLower(field)[0:9] == "services." {
			sfield := field[10:]
			for _, node := range nc.Nodes {
				NodeLoop:
				for _, service := range node.Services {
					if len(sfield) > 5 && strings.ToLower(field)[0:5] == "port." {
						ssfield := sfield[6:]
						if matchInInterface(&service.Port, ssfield, filter) {
							out = append(out, &node)
							break NodeLoop
						}

					} else if len(sfield) > 9 && strings.ToLower(field)[0:9] == "commands." {
						ssfield := sfield[10:]
						for _, command := range service.Commands {
							if matchInInterface(&command, ssfield, filter) {
								out = append(out, &node)
								break NodeLoop
							}
						}
					}
				}
			}
		} else if len(field) > 6 && strings.ToLower(field)[0:6] == "ports." {
			sfield := field[7:]
			for _, node := range nc.Nodes {
				NodeLoop2:
				for _, port := range node.Ports {
					if matchInInterface(&port, sfield, filter) {
						out = append(out, &node)
						break NodeLoop2
					}
				}
			}
		} else  if len(field) > 5  && strings.ToLower(field)[0:5] == "info." {
			sfield := field[6:]
			for _, node := range nc.Nodes {
				if matchInInterface(node.Info, sfield, filter) {
					out = append(out, &node)
				}
			}
		}else {
			return out, errors.New(fmt.Sprintf("Field doesn't start with 'nodes': <%s>", field))
		}
	} else {
		for _, node := range nc.Nodes {
			if matchInInterface(&node, field, filter) {
					out = append(out, &node)
			}
		}
	}
	if err != nil {
		return nil, errors.New(fmt.Sprintf("DiscoverReporter.Recover - Error: %s", err))
	}
	return out, err
}

func matchInInterface(itf interface{}, field string, filter regexp.Regexp) bool {
	var out bool = false
	if value := fieldValue(itf, field); value.IsValid() && ! value.IsNil() {
		return filter.Match([]byte(value.String()))
	}
	return out
}

func fieldValue(v interface{}, field string) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(v)).FieldByName(field)
}

func (nc *nodeCache) List() []types.Node {
	defer func(){
		if r := recover(); r != nil {
			fmt.Printf("DiscoverReporter.List - Error: %v\n", r)
		}
		nc.Unlock()
	}()
	nc.Lock()
	return nc.Nodes
}


func (nc *nodeCache) load() error {
	if "" == nc.FilePath {
		return errors.New("DiscoverReporter.load - Empty file ...")
	}
	var err error
	if _,err := os.Stat(nc.FilePath); err != nil{
		if err != nil {
			return errors.New(fmt.Sprintf("DiscoverReporter.load - (Registry file doesn't exist) Error: %s", err))
		}
	}
	file, errF := os.Open(nc.FilePath)
	if errF != nil{
		if err != nil {
			return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
		}
	}
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %v", r))
		}
		file.Close()
	}()
	var encoded []byte
	encoded, err = ioutil.ReadAll(file)
	if err != nil{
		if err != nil {
			return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
		}
	}
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(encoded))
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	count, err := decoder.Read(dbuf)
	if err != nil && err != io.EOF {
		return errors.New(fmt.Sprintf("DiscoverReporter.load - Decoder Read failed: %s", err))
	}
	if count > 0 {
		var out []types.Node = make([]types.Node, 0)
		itf, err := cio.Unmashall(dbuf, out, nc.Encoding)
		if err != nil{
			return errors.New(fmt.Sprintf("DiscoverReporter.load - (Decoding Issues) Error: %s", err))
		}
		nc.Nodes = itf.([]types.Node)
	} else {
		nc.Nodes = make([]types.Node, 0)
	}
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
	}
	return err
}

func (nc *nodeCache) save() error {
	if "" == nc.FilePath {
		return errors.New("DiscoverReporter.save - Empty file ...")
	}
	var exists bool = cio.ExistsFile(nc.FilePath)
	if ! exists {
		cio.CreateFileFolders(nc.FilePath, types.DefaultFolderPerm)
	}
	dt, err := cio.Marshall(nc.Nodes, nc.Encoding)
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.save - (Encoding Issues) Error: %s", err))
	}
	data := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, data)
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.save - Error: %v", r))
		}
		encoder.Close()
	}()
	_, err = encoder.Write([]byte(dt))
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.save - Encoder Write failed: %s", err))
	}
	if exists {
		cio.DeleteOrTruncateFile(nc.FilePath)
	}
	err = ioutil.WriteFile(nc.FilePath, data.Bytes(), types.DefaultFilePerm)
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.save - Error: %s", err))
	}
	return err
}

func NewClusterRegistryWithInternal(file string, encoding	cio.ParserFormat) ClusterRegistry {
	return &nodeCache{
		Nodes:    make([]types.Node, 0),
		Encoding: encoding,
		FilePath: file,
	}
}
func NewClusterRegistry(file string) ClusterRegistry {
	return &nodeCache{
		Nodes:    make([]types.Node, 0),
		Encoding: cio.ParserFormatYaml,
		FilePath: file,
	}
}

func NewInMemoryClusterRegistry() ClusterRegistry {
	return &nodeCache{
		Nodes:    make([]types.Node, 0),
		Encoding: cio.ParserFormat("NONE"),
		FilePath: "",
	}
}