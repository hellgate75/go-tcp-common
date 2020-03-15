package server

import (
	"crypto/rand"
	"crypto/tls"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"net"
	"net/http"
	"sync"
	"time"
)

var DEFAULT_ROOT_PATH_CALLBACK = func(w http.ResponseWriter, req *http.Request)() {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Welcome to the rest server\n"))
}

type TLSHandleFunc func(*tls.Conn, common.RestServer)()

type restServer struct {
	sync.RWMutex
	http.ServeMux
	server     *http.Server
	config      *tls.Config
	paths       map[string]*common.HandlerStruct
	tlsMode     bool
	logger      log.Logger
	handlerFunc TLSHandleFunc
	listener	*net.Listener
	conn		[]*tls.Conn
}

var (
	DEFAULT_HEADER_READ_TIMEOUT time.Duration = 60 * time.Second
	DEFAULT_READ_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_WRITE_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_IDLE_TIMEOUT time.Duration = 600 * time.Second
)

func New(logger log.Logger) common.RestServer {
	return NewHandleFunc(nil, logger)
}


func NewHandleFunc(handleFunc TLSHandleFunc, logger log.Logger) common.RestServer {
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
		Renegotiation: tls.RenegotiateNever,
	}
	return &restServer {
		config:     	tlsCfg,
		server:     	nil,
		paths:      	make(map[string]*common.HandlerStruct),
		tlsMode:    	false,
		logger:     	logger,
		handlerFunc: 	handleFunc,
		conn: 			make([]*tls.Conn, 0),
		listener: 		nil,
	}
}