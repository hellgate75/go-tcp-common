package common

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Security struct {
	CaCertificate   string
	CertFile        string
	KeyFile         string
	UseTLS          bool
}

type ApiModelStream interface{
	Parse() ApiModelStream
	FilterByKey(filter *KeyFiler) ApiModelStream
	GetAll() map[string]interface{}
}

type ApiStream interface {
	Read(r *io.Reader) (ApiStream, error)
	UseModel(model  *Model) ApiModelStream
	FilterContent(re *regexp.Regexp, splitSeparator string) ApiStream
	ReplaceInContent(re *regexp.Regexp, splitSeparator string, newValue string) ApiStream
	Stream() ApiStream
	Output(w http.ResponseWriter) error
}

type apiStream struct{
	_m              *Model
	_buf            *bytes.Buffer
	_objMap         map[string]interface{}
}

func (as *apiStream) Read(r *io.Reader) (ApiStream, error) {
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

func (as *apiStream) UseModel(model  *Model) ApiModelStream {
	as._m = model
	return as
}
func (as *apiStream) Parse() ApiModelStream {
	return as
}
func (as *apiStream) GetAll() map[string]interface{} {
	return make(map[string]interface{})
}

func (as *apiStream) FilterByKey(filter *KeyFiler) ApiModelStream {
	if filter != nil {
	
	}
	return as
}
func (as *apiStream) FilterContent(re *regexp.Regexp, splitSeparator string) ApiStream {
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
func (as *apiStream) ReplaceInContent(re *regexp.Regexp, splitSeparator string, newValue string) ApiStream {
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

func (as *apiStream) Stream() ApiStream {
	return as
}
func (as *apiStream) Output(w http.ResponseWriter) error {
	_, err := as._buf.WriteTo(w)
	return err
}

func NewApiStream() ApiStream{
	buff := bytes.NewBuffer([]byte{})
	return &apiStream{
		_buf: buff,
		_objMap: make(map[string]interface{}),
		_m: nil,
		
	}
}