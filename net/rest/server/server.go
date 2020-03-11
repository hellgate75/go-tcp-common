package server

import (
	"crypto/rand"
	"crypto/tls"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"net/http"
	"sync"
	"time"
)

//type RestHandler func(w http.ResponseWriter, req *http.Request)()
type RestCallback func(w http.ResponseWriter, req *http.Request, path string, accepts common.MimeType, produces common.MimeType)()

var DEFAULT_ROOT_PATH_CALLBACK = func(w http.ResponseWriter, req *http.Request)() {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Welcome to the rest server\n"))
}

//var funcEncapsulator = func(callBack *RestCallback, accepts *common.MimeType, produces *common.MimeType) RestHandler {
//	return func(w http.ResponseWriter, req *http.Request)(){
//		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
//		if accepts != nil {
//			w.Header().Set("Accept", string(*accepts))
//		}
//		if produces != nil {
//			w.Header().Set("Content-Type", string(*produces))
//		}
//		(*callBack)(w, req)
//	}
//}

type RestServer interface {
	AddPath(path string, callback RestCallback, accepts *common.MimeType, produces *common.MimeType) bool
	AddRootPath(callback RestCallback, accepts *common.MimeType, produces *common.MimeType) bool
	StartTLS(hostOrIpAddress string, port int32, cert string, key string) error
	Start(hostOrIpAddress string, port int32) error
	Stop() error
	IsRunning() bool
	WaitFor() error
}

type HandlerStruct struct {
	Handler     *RestCallback
	Consumes     *common.MimeType
	Produces     *common.MimeType
	Path         string
}

type restServer struct {
	sync.RWMutex
	http.ServeMux
	server     *http.Server
	config      *tls.Config
	paths       map[string]*HandlerStruct
	tlsMode     bool
	logger      log.Logger
}

var (
	DEFAULT_HEADER_READ_TIMEOUT time.Duration = 60 * time.Second
	DEFAULT_READ_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_WRITE_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_IDLE_TIMEOUT time.Duration = 600 * time.Second
)

func New(logger log.Logger) RestServer {
	return NewInsecure(false, logger)
}

func NewInsecure(allowInsecureConnections bool, logger log.Logger) RestServer {
	tlsCfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		ClientSessionCache: tls.NewLRUClientSessionCache(256),
		Rand: rand.Reader,
		InsecureSkipVerify: allowInsecureConnections,
		Renegotiation: tls.RenegotiateNever,
	}
	return &restServer {
		config:     tlsCfg,
		server:     nil,
		paths:      make(map[string]*HandlerStruct),
		tlsMode:    false,
		logger:     logger,
	}
}