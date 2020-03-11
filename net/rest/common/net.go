package common

import (
	"fmt"
	"net/http"
	"net/url"
)

// Key-Value Pair structure that contains TLS Client/Server Certificate and Key file path pair
type CertificateKeyPair struct {
	Cert string
	Key  string
}

// Generic Rest Client interface
type RestClient interface {
	Open() error
	Request(protocol RestProtocol, path string, method RestMethod, accepts *MimeType, body *[]byte, values *url.Values) (int, []byte, error)
}

// Generic Rest Callback function for handling pattern request
type RestCallback func(w http.ResponseWriter, req *http.Request, path string, accepts MimeType, produces MimeType) ()

// Generic Rest Server interface
type RestServer interface {
	AddPath(path string, callback RestCallback, accepts *MimeType, produces *MimeType, allowedMethods []RestMethod) bool
	AddRootPath(callback RestCallback, accepts *MimeType, produces *MimeType, allowedMethods []RestMethod) bool
	StartTLS(hostOrIpAddress string, port int32, cert string, key string) error
	Start(hostOrIpAddress string, port int32) error
	Stop() error
	IsRunning() bool
	WaitFor() error
}

// Structure containing 
type HandlerStruct struct {
	Handler  *RestCallback
	Consumes *MimeType
	Produces *MimeType
	Path     string
	Methods  []RestMethod
}

func (hs HandlerStruct) String() string {
	return fmt.Sprintf("HandlerStruct{Handler: %v, Comsumes: %s, Produces: %s, Path: %s, Web Methods: %v}",
		hs.Handler != nil, *hs.Consumes, *hs.Produces, hs.Path, hs.Methods)
}

