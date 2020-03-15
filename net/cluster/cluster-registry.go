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
	filePath	string
	persist		bool
	nodes		[]types.Node
	encoding	cio.ParserFormat
}

func (nc *nodeCache) Register(n *types.Node) error {
	if n == nil {
		return errors.New("DiscoverReporter.Register - Nil node reference")
	}
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.Register - Error: %v", r))
		}
		nc.Unlock()
	}()
	nc.Lock()
	nc.nodes = append(nc.nodes, *n)
	if nc.persist {
		err = nc.save()
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
	locked = true
	nc.Lock()
	if len(nodes) > 0 {
		for _, node := range nodes {
			node.Update(&n)
		}
		if nc.persist {
			err = nc.save()
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
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.Recover - Error: %v", r))
		}
		nc.Unlock()
	}()
	if nc.persist {
		err = nc.load()
		if err != nil {
			return out, errors.New(fmt.Sprintf("DiscoverReporter.Recover - Error: %s", err))
		}
	}
	if strings.Index(field, ".") > 0{
		if len(field) > 9 && strings.ToLower(field)[0:9] == "services." {
			sfield := field[10:]
			for _, node := range nc.nodes {
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
			for _, node := range nc.nodes {
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
			for _, node := range nc.nodes {
				if matchInInterface(node.Info, sfield, filter) {
					out = append(out, &node)
				}
			}
		}else {
			return out, errors.New(fmt.Sprintf("Field doesn't start with 'nodes': <%s>", field))
		}
	} else {
		for _, node := range nc.nodes {
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
	return nc.nodes
}


func (nc *nodeCache) load() error {
	if "" == nc.filePath {
		return errors.New("DiscoverReporter.load - Empty file ...")
	}
	var err error
	if _,err := os.Stat(nc.filePath); err != nil{
		if err != nil {
			return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
		}
	}
	file, errF := os.Open(nc.filePath)
	if errF != nil{
		if err != nil {
			return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
		}
	}
	var encoded []byte
	encoded, err = ioutil.ReadAll(file)
	if err != nil{
		if err != nil {
			return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
		}
	}
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(encoded))
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %v", r))
		}
	}()
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	count, err := decoder.Read(dbuf)
	if err != nil && err != io.EOF {
		return errors.New(fmt.Sprintf("DiscoverReporter.load - Decoder Read failed: %s", err))
	}
	if count > 0 {
		var out []types.Node = make([]types.Node, 0)
		itf, err := cio.Unmashall(dbuf, out, nc.encoding)
		if err != nil{
			if err != nil {
				return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
			}
		}
		nc.nodes = itf.([]types.Node)
	} else {
		nc.nodes = make([]types.Node, 0)
	}
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.load - Error: %s", err))
	}
	return err
}

func (nc *nodeCache) save() error {
	if "" == nc.filePath {
		return errors.New("DiscoverReporter.save - Empty file ...")
	}
	dt, err := cio.Marshall(nc.nodes, nc.encoding)
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.save - Error: %s", err))
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
	if _,err = os.Stat(nc.filePath); err==nil{
		os.Remove(nc.filePath)
	}
	err = ioutil.WriteFile(nc.filePath, data.Bytes(), 0660)
	if err != nil {
		return errors.New(fmt.Sprintf("DiscoverReporter.save - Error: %s", err))
	}
	return err
}

func NewClusterRegistryWithInternal(file string, encoding	cio.ParserFormat) ClusterRegistry {
	return &nodeCache{
		nodes: make([]types.Node, 0),
		encoding: encoding,
		filePath: file,
		persist: true,
	}
}
func NewClusterRegistry(file string) ClusterRegistry {
	return &nodeCache{
		nodes: make([]types.Node, 0),
		encoding: cio.ParserFormatYaml,
		filePath: file,
		persist: true,
	}
}

func NewInMemoryClusterRegistry() ClusterRegistry {
	return &nodeCache{
		nodes: make([]types.Node, 0),
		encoding: cio.ParserFormat("NONE"),
		filePath: "",
		persist: false,
	}
}