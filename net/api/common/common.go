package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	cio "github.com/hellgate75/go-tcp-common/io"

)

type Security struct {
	CaCertificate   string
	CertFile        string
	KeyFile         string
	UseTLS          bool
}

// Contains information about map key filtering expression
type KeyFiler struct{
	Key             string
	RegExp          regexp.Regexp
}

// ModelDataStream Describes how map an api data flow in object
type ModelDataStream interface{
	// Parse content data using one of available parsers: Yaml, Json, Xml
	Parse() ModelDataStream
	// Filter elements in the data list using the key expression
	FilterByKey(filter *KeyFiler) ModelDataStream
	// Retrive data from the applied filters
	GetAll() []map[string]interface{}
}

// Defines the way data can be loaded and filtered
type DataStream interface {
	// Return the Model Data Stream connected to this data stream
	ToModel(parserType  cio.ParserFormat) ModelDataStream
	// Filter data content via RegExp, suing filter split block separator string
	FilterContent(re *regexp.Regexp, splitSeparator string) DataStream
	// Replace data content via RegExp, suing filter split block separator string
	ReplaceInContent(re *regexp.Regexp, splitSeparator string, newValue string) DataStream
	// Read from the input stream reader, if applicable
	Fetch() (int64, DataStream)
	// Retrieves information about availability to fetch data
	CanFetch() bool
	// Write data to uotput writer anc reset/clear content
	Output(w http.ResponseWriter) error
}

type apiStream struct{
	_buf            *bytes.Buffer
	_objMap         []map[string]interface{}
	_type           *cio.ParserFormat
	_r              *io.Reader
}

func (as *apiStream) CanFetch() bool {
	return as._r != nil
}


func (as *apiStream) Fetch() (int64, DataStream) {
	if as._r != nil {
		n, err := as._buf.ReadFrom(*as._r)
		if err == nil {
			return n, as
		}
	}
	return 0, as
}

func (as *apiStream) Read(r *io.Reader) (DataStream, error) {
	if r == nil {
		return as, errors.New("Null input stream reader")
	}
	var err error = nil
	as._buf.Reset()
	_, err = as._buf.ReadFrom(*r)
	return as, err
}

func split(s []byte, sep string) []string{
	var out []string = make([]string, 0)
	out = append(out, strings.Split(string(s), sep)...)
	return out
}

func merge(s []string, sep string) []byte{
	var out string = ""
	for _,sn := range s {
		sepX := sep
		if len(out) == 0 {
			sepX = ""
		}
		out += sepX + sn
	}
	return []byte(out)
}

type modelStruct []map[string]interface{}

func (as *apiStream) ToModel(parserType  cio.ParserFormat) ModelDataStream {
	as._type = &parserType
	return as
}

func (as *apiStream) Parse() ModelDataStream {
	if as._type == nil {
		return as
	}
	var m []map[string]interface{} = make([]map[string]interface{}, 0)
	var model modelStruct = m
	var data []byte = as._buf.Bytes()
	itf, err := cio.Unmashall(data, model, *as._type)
	if err != nil {
		return as
	}
	if mdl, ok := itf.(modelStruct); ok {
		as._objMap = mdl
	}
	return as
}

func (as *apiStream) GetAll() []map[string]interface{} {
	return as._objMap
}

func (as *apiStream) FilterByKey(filter *KeyFiler) ModelDataStream {
	if filter != nil {
		var list []map[string]interface{} = make([]map[string]interface{}, 0)
		for _,item := range as._objMap {
			if value, ok := item[filter.Key]; ok {
				if fmt.Sprintf("%T", value) == "string" {
					if filter.RegExp.MatchString(fmt.Sprintf("%s", value)) {
						list = append(list, item)
					}
				} else {
					list = append(list, item)
				}
			} else {
				list = append(list, item)
			}
		}
		as._objMap = list
	}
	return as
}
func (as *apiStream) FilterContent(re *regexp.Regexp, splitSeparator string) DataStream {
	if re != nil {
		var out []string = make([]string, 0)
		for _, s := range split(as._buf.Bytes(), splitSeparator) {
			if len(re.FindAllStringIndex(s, 0)) > 0 {
				out = append(out, s)
			}
		}
		as._buf.Reset()
		as._buf.Read(merge(out, splitSeparator))
	}
	return as
}

func (as *apiStream) ReplaceInContent(re *regexp.Regexp, splitSeparator string, newValue string) DataStream {
	if re != nil {
		var out []string = make([]string, 0)
		for _, s := range split(as._buf.Bytes(), splitSeparator) {
			if len(re.FindAllStringIndex(s, 0)) > 0 {
				s = re.ReplaceAllString(s, newValue)
			}
			out = append(out, s)
		}
		as._buf.Reset()
		as._buf.Read(merge(out, splitSeparator))
	}
	return as
}

func (as *apiStream) Output(w http.ResponseWriter) error {
	_, err := as._buf.WriteTo(w)
	return err
}

// Create a new stream from array of bytes
func NewStream(data []byte) (DataStream, error){
	if len(data) == 0 {
		return nil, errors.New("Empty Stream buffer")
	}
	buff := bytes.NewBuffer(data)
	return &apiStream{
		_buf: buff,
		_objMap: make([]map[string]interface{}, 0),
		_type: nil,
		_r: nil,
	}, nil
}

// Create a new stream from limited or continuous input io.Reader
func NewStreamFrom(r *io.Reader) (DataStream, error){
	if r == nil {
		return nil, errors.New("Nil stream reader in imput")
	}
	buff := bytes.NewBuffer([]byte{})
	n, err := buff.ReadFrom(*r)
	if err != nil {
		return nil, err
	}
	if n == 0 || buff.Len() == 0 {
		return nil, errors.New("Empty Stream buffer")
	}
	return (&apiStream{
		_buf: buff,
		_objMap: make([]map[string]interface{}, 0),
		_type: nil,
		_r: r,
	}).Read(r)
}

// Create a new stream from limited input io.File
func NewFileStream(f *os.File) (DataStream, error){
	if f == nil {
		return nil, errors.New("Nil file in imput")
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New(fmt.Sprintf("Unable to load data from file: %s", f.Name()))
	}
	buff := bytes.NewBuffer(data)
	return &apiStream{
		_buf: buff,
		_objMap: make([]map[string]interface{}, 0),
		_type: nil,
		_r: nil,
	}, nil
}

// Create a new stream from continuous input io.PipeReader (from Pipe)
func NewPipeStream(r *io.PipeReader) (DataStream, error){
	if r == nil {
		return nil, errors.New("Nil pipe reader in imput")
	}
	buff := bytes.NewBuffer([]byte{})
	_, err := buff.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	var reader io.Reader = r
	return &apiStream{
		_buf: buff,
		_objMap: make([]map[string]interface{}, 0),
		_type: nil,
		_r: &reader,
	}, nil
}

func filesToBytes(filePath string, recursive bool) []byte {
	if len(filePath) == 0 {
		return []byte{}
	}
	var out = bytes.NewBuffer([]byte{})
	if fi, err := os.Stat(filePath); err != nil {
		if fi.IsDir() {
			fs, err := ioutil.ReadDir(filePath)
			if err == nil {
				for _, fItem := range fs {
					var name string = filePath + string(os.PathSeparator) + fItem.Name()
					if fItem.IsDir() {
						if recursive {
							out.Read(filesToBytes(name, recursive))
						}
					} else {
						if file, err := os.Open(name); err == nil {
							data, err := ioutil.ReadAll(file)
							if err == nil {
								out.Read(data)
							}
						}
					}
				}
			}
			
		} else {
			if file, err := os.Open(filePath); err == nil {
				data, err := ioutil.ReadAll(file)
				if err == nil {
					out.Read(data)
				}
			}
		}
		
	}
	return out.Bytes()
}

// Create a new stream from bytes of files folder in a directory, eventually crowling for sub directories
func NewOlderContentStream(folder string, recursive bool) (DataStream, error){
	data := filesToBytes(folder, recursive)
	if len(data) == 0 {
		return nil, errors.New(fmt.Sprintf("Unable to load data from folder: '%s'", folder))
	}
	buff := bytes.NewBuffer(data)
	return &apiStream{
		_buf: buff,
		_objMap: make([]map[string]interface{}, 0),
		_type: nil,
		_r: nil,
	}, nil
}


// Create a new stream from limited input as output by execution of the given command and arguments
func NewCommandStream(command string, args ...string) (DataStream, error){
	if len(command) == 0{
		return nil, errors.New("Empty command ...")
	}
	var cmd *exec.Cmd = nil
	if len(args) > 0 {
		cmd = exec.Command(command, args...)
	} else {
		cmd = exec.Command(command)
	}
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	dataStr := fmt.Sprintf("%s", stdoutStderr)
	
	if len(dataStr) == 0 {
		return nil, errors.New(fmt.Sprintf("Unable to load data for command: %v", command))
	}
	buff := bytes.NewBuffer([]byte(dataStr))
	
	return &apiStream{
		_buf: buff,
		_objMap: make([]map[string]interface{}, 0),
		_type: nil,
		_r: nil,
	}, nil
}

// Interface that describes the callback action of an API call
type ApiAction interface{
	// Execute API command with API given arguments
	Run(Args ...interface{}) error
}