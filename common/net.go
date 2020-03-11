package common

import (
	"fmt"
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"net/http"
	"net/url"
)

type CertificateKeyPair struct {
	Cert string
	Key  string
}

type RestClient interface {
	Open() error
	Request(protocol common.RestProtocol, path string, method common.RestMethod, accepts *common.MimeType, body *[]byte, values *url.Values) (int, []byte, error)
}

//type RestHandler func(w http.ResponseWriter, req *http.Request)()
type RestCallback func(w http.ResponseWriter, req *http.Request, path string, accepts common.MimeType, produces common.MimeType) ()

type RestServer interface {
	AddPath(path string, callback RestCallback, accepts *common.MimeType, produces *common.MimeType, allowedMethods []common.RestMethod) bool
	AddRootPath(callback RestCallback, accepts *common.MimeType, produces *common.MimeType, allowedMethods []common.RestMethod) bool
	StartTLS(hostOrIpAddress string, port int32, cert string, key string) error
	Start(hostOrIpAddress string, port int32) error
	Stop() error
	IsRunning() bool
	WaitFor() error
}

type HandlerStruct struct {
	Handler  *RestCallback
	Consumes *common.MimeType
	Produces *common.MimeType
	Path     string
	Methods  []common.RestMethod
}

func (hs HandlerStruct) String() string {
	return fmt.Sprintf("HandlerStruct{Handler: %v, Comsumes: %s, Produces: %s, Path: %s, Web Methods: %v}",
		hs.Handler != nil, *hs.Consumes, *hs.Produces, hs.Path, hs.Methods)
}

