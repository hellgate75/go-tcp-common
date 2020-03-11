package server

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	common2 "github.com/hellgate75/go-tcp-common/common"
	"github.com/hellgate75/go-tcp-common/log"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

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

type restServer struct {
	sync.RWMutex
	http.ServeMux
	CaCert     string
	server     *http.Server
	config      *tls.Config
	paths       map[string]*common2.HandlerStruct
	tlsMode     bool
	logger      log.Logger
}

var (
	DEFAULT_HEADER_READ_TIMEOUT time.Duration = 60 * time.Second
	DEFAULT_READ_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_WRITE_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_IDLE_TIMEOUT time.Duration = 600 * time.Second
)

func New(logger log.Logger) common2.RestServer {
	return NewCaCert(false, "", logger)
}

func NewCaCert(allowInsecureConnections bool, caCert string, logger log.Logger) common2.RestServer {
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
	if "" != caCert {
		if logger != nil {
			logger.Errorf("client: using ca cert: <%s>", caCert)
		}
		caCert, err := ioutil.ReadFile(caCert)
		if err != nil {
			if logger != nil {
				logger.Errorf("client: using ca cert: <%s>, details: %s", caCert, err.Error())
			}
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				if logger != nil {
					logger.Error("No certs appended, using system certificates only")
					logger.Fatalf("client: loadkeys: %s", err.Error())
				}
			} else {
				tlsCfg.RootCAs= caCertPool
			}
		}
	}
	return &restServer {
		config:     tlsCfg,
		server:     nil,
		paths:      make(map[string]*common2.HandlerStruct),
		tlsMode:    false,
		logger:     logger,
		CaCert:     caCert,
	}
}