package common

import (
	"fmt"
	"github.com/hellgate75/go-tcp-common/io/streams"
	"github.com/hellgate75/go-tcp-common/net/common"
	"net/url"
)

type TLSConfig struct {
	CaCertificate   string
	CertFile        string
	KeyFile         string
	UseInsecure     bool
}

// Interface that describes the callback action of an API call
type ApiAction interface{
	// Execute API command with API given arguments
	Run(Args ...interface{}) error
}

type ApiServer interface{
	Start(ipAddress string, port int64) error
	StartTLS(ipAddress string, port int64, config *TLSConfig) error
	Stop() error
	Shutdown() error
	IsRunning() bool
	AddApiAction(path string, action ApiAction, hasInternalAnswer bool, method *common.RestMethod, produces *common.MimeType, consumes *common.MimeType) bool
	AddApiStream(path string, stream streams.DataStream, method *common.RestMethod, produces *common.MimeType, consumes *common.MimeType) bool
}

type APIClient interface {
	Connect(ipAddress string, port int64) error
	ConnectTSL(ipAddress string, port int64, config *TLSConfig) error
	Close() error
	GetApi(protocol common.RestProtocol,path string, method *common.RestMethod, produces *common.MimeType, consumes *common.MimeType, body *[]byte, values *url.Values) (int, []byte, error)
}

type HandlerRef struct{
	Path        string
	Action      ApiAction
	HasAnswer   bool
	Stream      streams.DataStream
	Method      *common.RestMethod
	Consumes     *common.MimeType
	Produces    *common.MimeType
}

func (ha *HandlerRef) String() string {
	return fmt.Sprintf("HandlerRef{Path: \"%s\", Action: %v, Stream: %v, Method: %v, Produces: %v, Consumes: %v}",
		ha.Path, ha.Action != nil, ha.Stream != nil, *ha.Method, *ha.Produces, *ha.Consumes)
}

func (ha *HandlerRef) IsAction() bool {
	return ha.Action != nil
}

func (ha *HandlerRef) IsStream() bool {
	return ha.Stream != nil
}