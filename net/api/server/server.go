package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hellgate75/go-tcp-common/io"
	"github.com/hellgate75/go-tcp-common/io/streams"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/api/common"
	ncom "github.com/hellgate75/go-tcp-common/net/common"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

type apiServer struct{
	sync.Mutex
	router          *mux.Router
	logger          log.Logger
	server          *http.Server
	routes          map[string]*common.HandlerRef
}
var (
	DEFAULT_HEADER_READ_TIMEOUT time.Duration = 60 * time.Second
	DEFAULT_READ_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_WRITE_TIMEOUT time.Duration = 600 * time.Second
	DEFAULT_IDLE_TIMEOUT time.Duration = 600 * time.Second
)

func (as *apiServer) StartTLS(ipAddress string, port int64, config *common.TLSConfig) error {
	if as.server != nil {
		return errors.New("Server already running!!")
	}
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
		InsecureSkipVerify: config.UseInsecure,
		Renegotiation: tls.RenegotiateNever,
	}
	if "" != config.CaCertificate {
		if as.logger != nil {
			as.logger.Debugf("api: server: using ca cert: <%s>", config.CaCertificate)
		}
		caCert, err := ioutil.ReadFile(config.CaCertificate)
		if err != nil {
			if as.logger != nil {
				as.logger.Errorf("api: server: using ca cert: <%s>, details: %s", caCert, err.Error())
			}
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				if as.logger != nil {
					as.logger.Error("No certs appended, using system certificates only")
					as.logger.Fatalf("api: server: loadkeys: %s", err.Error())
				}
			} else {
				tlsCfg.RootCAs= caCertPool
			}
		}
	}
	if "" != config.KeyFile &&  "" != config.CertFile {
		as.logger.Debugf("api: server: using server key: <%s>, cert: <%s> ", config.KeyFile, config.CertFile)
		as.logger.Debugf("api: server: using server key: <%s>, cert: <%s> ", config.KeyFile, config.CertFile)
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			as.logger.Errorf("api: server: Unable to load key : %s and certificate: %s", config.KeyFile, config.CertFile)
			as.logger.Fatalf("api: server: loadkeys: %s", err.Error())
		} else {
			tlsCfg.Certificates=[]tls.Certificate{cert}
		}
	}
	as.server = &http.Server{
		Addr: fmt.Sprintf("%s:%v", ipAddress, port),
		TLSConfig: tlsCfg,
		Handler: as.router,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context{
			ctx = context.WithValue(ctx, ncom.ContextRemoteAddress, c.RemoteAddr())
			sessionKey, err := uuid.NewV4()
			if err == nil {
				ctx = context.WithValue(ctx, ncom.ContextSessionKey, sessionKey)
			} else {
				if as.logger != nil {
					as.logger.Errorf("api: server: start : tls: Error retriving session id, Details: %s", err)
				}
			}
			ctx = context.WithValue(ctx, ncom.ContextKeyAuthtoken, ncom.GenerateSecureToken(64))
			return ctx
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		ReadTimeout: DEFAULT_READ_TIMEOUT,
		ReadHeaderTimeout: DEFAULT_HEADER_READ_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
		IdleTimeout: DEFAULT_IDLE_TIMEOUT,
	}
	return as.server.ListenAndServe()
}
func (as *apiServer) Start(ipAddress string, port int64) error {
	if as.server != nil {
		return errors.New("Server already running!!")
	}
	as.server = &http.Server{
		Addr: fmt.Sprintf("%s:%v", ipAddress, port),
		Handler: as.router,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context{
			ctx = context.WithValue(ctx, ncom.ContextRemoteAddress, c.RemoteAddr())
			sessionKey, err := uuid.NewV4()
			if err == nil {
				ctx = context.WithValue(ctx, ncom.ContextSessionKey, sessionKey)
			} else {
				if as.logger != nil {
					as.logger.Errorf("api: server: start : simple: Error retriving session id, Details: %s", err)
				}
			}
			ctx = context.WithValue(ctx, ncom.ContextKeyAuthtoken, ncom.GenerateSecureToken(64))
			return ctx
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		ReadTimeout: DEFAULT_READ_TIMEOUT,
		ReadHeaderTimeout: DEFAULT_HEADER_READ_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
		IdleTimeout: DEFAULT_IDLE_TIMEOUT,
	}
	return as.server.ListenAndServe()
}
func (as *apiServer) Shutdown() error {
	if as.server == nil {
		return nil
	}
	defer func() {
		as.server = nil
	}()
	return as.server.Shutdown(context.Background())
}

func (as *apiServer) Stop() error {
	if as.server == nil {
		return nil
	}
	defer func() {
		as.server = nil
	}()
	return as.server.Close()
}
func (as *apiServer) IsRunning() bool {
	return as.server != nil
}
func (as *apiServer) handle(w http.ResponseWriter, req *http.Request)(){
	path := req.URL.Path
	//method := ncom.RestMethod(req.Method)
	if handlerStruct, ok := as.routes[path]; ok {

		var requiredWebMethod string = req.Method

		as.logger.Debugf("api: server: exec-path: Requested Method: %s", requiredWebMethod)
		as.logger.Debugf("api: server: exec-path: Available Path %s Methods: %s", path, handlerStruct.Method)

		as.logger.Warnf("api: server: exec-path: Client accepts: %v", req.Header.Get("Accept"))
		as.logger.Warnf("api: server: exec-path: Client content: %v", req.Header.Get("Content-Type"))
		//if handlerStruct.Produces == nil || handlerStruct.Produces !=
		as.logger.Warnf("api: server: exec-path: Calling path: %s, func: %s", path, handlerStruct != nil)
		if handlerStruct.IsAction() {
			err := handlerStruct.Action.Run(req, w, requiredWebMethod)
			var code int = http.StatusOK
			var status string = ""
			if err != nil {
				code = http.StatusInternalServerError
				message:=fmt.Sprintf("api: server: exec-path: Calling path: %s, details: %s", path, err)
				status = "status: KO\n"+message
				if string(*handlerStruct.Produces) == string(ncom.XML_MIME_TYPE) {
					status = "<status>KO</status>\n<message>"+message+"<message>"
				} else if string(*handlerStruct.Produces) == string(ncom.YAML_MIME_TYPE) {
					status = "status: KO\nmessage: "+message
				} else if string(*handlerStruct.Produces) == string(ncom.JSON_MIME_TYPE) {
					status = "{\n\"status\": \"KO\"\n\"message\": \""+message+"\"\n}"
				}
				ncom.SubmitFaiure(w, http.StatusInternalServerError, status)
			} else {
				message:=fmt.Sprintf("api: server: exec-path: Calling path: %s, status: %s", path, "OK")
				status = "status: KO\n"+message
				if string(*handlerStruct.Produces) == string(ncom.XML_MIME_TYPE) {
					status = "<status>OK</status>\n<message>"+message+"<message>"
				} else if string(*handlerStruct.Produces) == string(ncom.YAML_MIME_TYPE) {
					status = "status: OK\nmessage: "+message
				} else if string(*handlerStruct.Produces) == string(ncom.JSON_MIME_TYPE) {
					status = "{\n\"status\": \"OK\"\n\"message\": \""+message+"\"\n}"
				}
			}
			ncom.SubmitFaiure(w, code, status)
		} else if handlerStruct.IsStream() {
			if handlerStruct.Stream.CanFetch() {
				handlerStruct.Stream.Fetch()
			}
			if string(*handlerStruct.Consumes) == string(ncom.XML_MIME_TYPE) {
				model := handlerStruct.Stream.ToModel(io.ParserFormatXml)
				if string(*handlerStruct.Produces) == string(ncom.XML_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatXml)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (XML/XML), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else if string(*handlerStruct.Produces) == string(ncom.YAML_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatYaml)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (XML/YAML), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else if string(*handlerStruct.Produces) == string(ncom.JSON_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatJson)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (XML/JSON), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else {
					ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s, Unknown format: %s", path, string(*handlerStruct.Produces)))
				}
			} else if string(*handlerStruct.Consumes) == string(ncom.YAML_MIME_TYPE) {
				model := handlerStruct.Stream.ToModel(io.ParserFormatYaml)
				if string(*handlerStruct.Produces) == string(ncom.XML_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatXml)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (YAML/XML), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else if string(*handlerStruct.Produces) == string(ncom.YAML_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatYaml)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (YAML/YAML), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else if string(*handlerStruct.Produces) == string(ncom.JSON_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatJson)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (YAML/JSON), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else {
					ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s, Unknown format: %s", path, string(*handlerStruct.Produces)))
				}
			} else if string(*handlerStruct.Consumes) == string(ncom.JSON_MIME_TYPE) {
				model := handlerStruct.Stream.ToModel(io.ParserFormatJson)
				if string(*handlerStruct.Produces) == string(ncom.XML_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatXml)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (JSON/XML), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else if string(*handlerStruct.Produces) == string(ncom.YAML_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatYaml)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (JSON/YAML), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else if string(*handlerStruct.Produces) == string(ncom.JSON_MIME_TYPE) {
					list := model.GetAll()
					data, err := io.Marshall(list, io.ParserFormatJson)
					if err != nil {
						ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s (JSON/JSON), Details: %s", path, err))
					} else {
						ncom.SubmitSuccess(w, string(data))
					}
				} else {
					ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s, Unknown format: %s", path, string(*handlerStruct.Produces)))
				}
			} else {
				buff := bytes.NewBuffer([]byte{})
				err := handlerStruct.Stream.Output(buff)
				if err != nil {
					ncom.SubmitFaiure(w,http.StatusInternalServerError, fmt.Sprintf("Errors path: %s, Details: %s", path, err))
				} else {
					ncom.SubmitSuccess(w, buff.String())
				}
			}
		} else {
			as.logger.Error("api: server: exec-path: No Action nor Stream available for execution")
		}
	} else {
		ncom.SubmitFaiure(w, http.StatusNotFound, "NOT_FOUND")
	}
}
func (as *apiServer) AddApiAction(path string, action common.ApiAction, method *ncom.RestMethod, produces *ncom.MimeType, consumes *ncom.MimeType) bool {
	out := false
	if _, ok := as.routes[path]; !ok {
		as.routes[path] = &common.HandlerRef{
			Method: method,
			Path: path,
			Action: action,
			Stream: nil,
			Produces: produces,
			Consumes: consumes,
		}
		as.router.HandleFunc(path, as.handle).Methods(string(*method))
		ok = true
	}
	
	return out
}
func (as *apiServer) AddApiStream(path string, stream streams.DataStream, method *ncom.RestMethod, produces *ncom.MimeType, consumes *ncom.MimeType) bool {
	out := false
	if _, ok := as.routes[path]; !ok {
		as.routes[path] = &common.HandlerRef{
			Method: method,
			Path: path,
			Action: nil,
			Stream: stream,
			Produces: produces,
			Consumes: consumes,
		}
		as.router.HandleFunc(path, as.handle).Methods(string(*method))
		ok = true
	}
	
	return out
}

func NewApiServer(logger log.Logger) common.ApiServer {
	return &apiServer{
		router: mux.NewRouter().StrictSlash(true),
		logger: logger,
		routes: make(map[string]*common.HandlerRef),
	}
}