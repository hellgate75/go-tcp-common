package common

import (
	"fmt"
	"github.com/hellgate75/go-tcp-common/net/common"
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
	// Authenticate and open connectivity with the server
	Open() error
	// Close connection and sign-off the server
	Close() error
	//Send a requerst to the connected server
	Request(protocol common.RestProtocol, path string, method common.RestMethod, accepts *common.MimeType, body *[]byte, values *url.Values) (int, []byte, error)
	// Returns information about server connectivity state
	IsConnected() bool
}

// Generic Rest Callback function for handling pattern request
type RestCallback func(w http.ResponseWriter, req *http.Request, path string, accepts common.MimeType, produces common.MimeType) ()

// Generic Rest Server interface
type RestServer interface {
	AddPath(path string, callback RestCallback, accepts *common.MimeType, produces *common.MimeType, allowedMethods []common.RestMethod) bool
	AddRootPath(callback RestCallback, accepts *common.MimeType, produces *common.MimeType, allowedMethods []common.RestMethod) bool
	StartTLS(hostOrIpAddress string, port int32, certs []CertificateKeyPair, CaCertificate string, insecure bool) error
	Start(hostOrIpAddress string, port int32) error
	Stop() error
	Shutdown() error
	IsRunning() bool
	WaitFor() error
}

// Structure containing 
type HandlerStruct struct {
	Handler  *RestCallback
	Consumes *common.MimeType
	Produces *common.MimeType
	Path     string
	Methods  []common.RestMethod
}

// String representation of the Handler Strcture
func (hs HandlerStruct) String() string {
	return fmt.Sprintf("HandlerStruct{Handler: %v, Comsumes: %s, Produces: %s, Path: %s, Web Methods: %v}",
		hs.Handler != nil, *hs.Consumes, *hs.Produces, hs.Path, hs.Methods)
}

