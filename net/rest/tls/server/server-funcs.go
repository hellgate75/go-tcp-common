package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	ncom "github.com/hellgate75/go-tcp-common/net/common"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func (rs *restServer) AddPath(path string, callback common.RestCallback, accepts *ncom.MimeType, produces *ncom.MimeType, allowedMethods []ncom.RestMethod) bool {
	var state bool = false
	defer func() {
		if r := recover(); r != nil {
			if rs.logger != nil {
				rs.logger.Errorf("server: add-path: Errors dueing TLS Rest Server Start up, Details: %v", r)
			}
			rs.RUnlock()
			state = false
		}
	}()
	if _, ok := rs.paths[path]; !ok {
		rs.RLock()
		handler := common.HandlerStruct{
			Handler: &callback,
			Consumes: accepts,
			Produces: produces,
			Path: path,
			Methods: allowedMethods,
		}
		rs.logger.Debugf("server: add-path: Adding Path: %s", handler)
		rs.paths[path] = &handler
		handlerFunc := func(w http.ResponseWriter, req *http.Request)() {
			defer func(){
				if r := recover(); r != nil {
					ncom.SubmitFaiure(w, http.StatusInternalServerError, fmt.Sprintf("Error: %v", r))
					return
				}
			}()
			var path string = req.URL.Path
			rs.logger.Debugf("server: exec-path: Requested Path: %s", path)
			if handlerStruct, ok := rs.paths[path]; ok {
				var matching bool = false
				var requiredWebMethod string = req.Method
				rs.logger.Debugf("server: exec-path: Requested Method: %s", requiredWebMethod)
				rs.logger.Debugf("server: exec-path: Available Path %s Methods: %v", path, handlerStruct.Methods)
				for _, wm := range handlerStruct.Methods {
					if string(wm) == requiredWebMethod {
						matching = true
					}
				}
				rs.logger.Debugf("server: exec-path: Fount in List: %v", matching)
				rs.logger.Warnf("server: exec-path: Client accepts: %v", req.Header.Get("Accept"))
				rs.logger.Warnf("server: exec-path: Client content: %v", req.Header.Get("Content-Type"))
				//if handlerStruct.Produces == nil || handlerStruct.Produces !=
				if ! matching {
					var message string = fmt.Sprintf("Web Method (path: %s): %s, not matching with available %v", path, requiredWebMethod, handlerStruct.Methods)
					rs.logger.Warnf(fmt.Sprintf("server: exec-path: Required " + message))
					ncom.SubmitFaiure(w, http.StatusMethodNotAllowed, message)
					return
				}
				rs.logger.Warnf("server: exec-path: Calling path: %s, func: %v", path, handlerStruct.Handler != nil)
				if handlerStruct.Handler != nil {
					(*handlerStruct.Handler)(w, req, path, *(*handlerStruct).Consumes, *(*handlerStruct).Consumes)
				} else {
					rs.logger.Warnf("server: exec-path: Unavailable Handler for path: %s", path)
				}
			} else {
				ncom.SubmitFaiure(w, http.StatusNotFound, "NOT_FOUND")
				return
			}
		}
		rs.HandleFunc(path, handlerFunc)
		rs.RUnlock()
		state = true
	}
	return state
}

func (rs *restServer) AddRootPath(callback common.RestCallback, accepts *ncom.MimeType, produces *ncom.MimeType, allowedMethods []ncom.RestMethod) bool {
	return rs.AddPath("/", callback, accepts, produces, allowedMethods)
}

func (rs *restServer) StartTLS(hostOrIpAddress string, port int32, certs []common.CertificateKeyPair, CaCertificate string, insecure bool) error {
	var err error = nil
	var locked bool = false
	defer func() {
		if r := recover(); r != nil {
			var message string =  fmt.Sprintf("server: start : tls: Errors during TLS Rest Server Start up, Details: %v", r)
			err = errors.New(message)
			if rs.logger != nil {
				rs.logger.Error(message)
			}
		} else {
			rs.tlsMode = true
		}
		if locked {
			rs.Unlock()
		}
	}()
	rs.Lock()
	locked = true
	if rs.server != nil {
		var mode string = "normal"
		if rs.tlsMode {
			mode = "TLS/TCP"
		}
		var message = fmt.Sprintf("server: start : tls: Server already started in %s mode!!", mode)
		if rs.logger != nil {
			rs.logger.Error(message)
		}
		return errors.New(fmt.Sprintf("server: start : tls: Server already started in %s mode!!", mode))
	}
	rs.config.InsecureSkipVerify = insecure
	if "" != CaCertificate {
		if rs.logger != nil {
			rs.logger.Debugf("server: using ca cert: <%s>", CaCertificate)
		}
		caCert, err := ioutil.ReadFile(CaCertificate)
		if err != nil {
			if rs.logger != nil {
				rs.logger.Errorf("server: using ca cert: <%s>, details: %s", caCert, err.Error())
			}
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				if rs.logger != nil {
					rs.logger.Error("No certs appended, using system certificates only")
					rs.logger.Fatalf("server: loadkeys: %s", err.Error())
				}
			} else {
				rs.config.RootCAs= caCertPool
			}
		}
	}
	if certs != nil && len(certs) > 0 {
		var crts []tls.Certificate = make([]tls.Certificate, 0)
		for _, config := range certs {
			if "" != config.Key &&  "" != config.Cert {
				rs.logger.Debugf("api: server: using server key: <%s>, cert: <%s> ", config.Key, config.Cert)
				rs.logger.Debugf("api: server: using server key: <%s>, cert: <%s> ", config.Key, config.Cert)
				cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
				if err != nil {
					rs.logger.Errorf("api: server: Unable to load key : %s and certificate: %s", config.Key, config.Cert)
					rs.logger.Fatalf("api: server: loadkeys: %s", err.Error())
				} else {
					crts = append(crts, cert)
				}
			}
		}
		rs.config.Certificates=crts
	}
	var cert string
	var key string
	if len(certs) > 0 {
		cert = certs[0].Cert
		key = certs[0].Key
	} else {
		cert = ""
		key = ""
	}

	if rs.handlerFunc == nil {
		rs.server = &http.Server{
			Addr: fmt.Sprintf("%s:%v", hostOrIpAddress, port),
			TLSConfig: rs.config,
			Handler: rs,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context{
				ctx = context.WithValue(ctx, ncom.ContextRemoteAddress, c.RemoteAddr())
				sessionKey, err := uuid.NewV4()
				if err == nil {
					ctx = context.WithValue(ctx, ncom.ContextSessionKey, sessionKey)
				} else {
					if rs.logger != nil {
						rs.logger.Errorf("server: start : tls: Error retriving session id, Details: %s", err)
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
		rs.Unlock()
		locked = false
		err = rs.server.ListenAndServeTLS(cert, key)
		if err != nil && "http: Server closed" != err.Error() {
			if rs.logger != nil {
				rs.logger.Errorf("server: start : tls: Error: %s", err)
			}
		}

	} else {
		if rs.logger != nil {
			rs.logger.Debugf("server: listen:  Using ip: %s, port: %v", hostOrIpAddress, port)
		}
		service := fmt.Sprintf("%s:%v",hostOrIpAddress, port)
		if rs.logger != nil {
			rs.logger.Debugf("server: listen:  Using address: %s", service)
		}

		var list net.Listener
		list, err = tls.Listen("tcp", service, rs.config)
		if err != nil {
			rs.logger.Fatalf("server: listen:  Error: %s", err)
			if rs.listener != nil {
				(*rs.listener).Close()
			}
			panic("server: listen: " + err.Error())
		}
		rs.listener = &list
		if rs.logger != nil {
			rs.logger.Infof("server: listen: %v", service)
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprintf("%v", r))
					rs.logger.Fatalf("TCP Server exit ...")
				}
				rs.logger.Info("TCP Server exit ...")
			}()
			for rs.IsRunning() {
				conn, errN := (*rs.listener).Accept()
				if errN != nil {
					rs.logger.Errorf("server: accept: %s", errN)
					continue
				}
				if rs.logger != nil {
					rs.logger.Debugf("server: accepted from %s", conn.RemoteAddr())
				}
				tlscon, ok := conn.(*tls.Conn)
				if ok {
					rs.logger.Debug("ok=true")
					state := tlscon.ConnectionState()
					for _, v := range state.PeerCertificates {
						rs.logger.Info(x509.MarshalPKIXPublicKey(v.PublicKey))
					}
				}
				rs.conn = append(rs.conn, tlscon)
				go func(tlsconn *tls.Conn, server common.RestServer, restStruct *restServer) {
					restStruct.handlerFunc(tlsconn, rs)
					tlsconn.Close()
					defer func() {
						if r := recover(); r != nil {
							if restStruct.logger != nil {
								restStruct.logger.Errorf("Error closing connection, error: %v", r)
							}
						}
						restStruct.Unlock()
					}()
					restStruct.Lock()
					conns := make([]*tls.Conn, 0)
					for _, c := range rs.conn {
						if c != tlsconn {
							conns = append(conns, c)
						}
					}
					restStruct.conn = conns
				}(tlscon, rs, rs)
			}
		}()
		rs.Unlock()
		locked = false
	}
	return err
}

func (rs *restServer) Start(hostOrIpAddress string, port int32) error {
	var err error = nil
	var locked bool = false
	defer func() {
		if r := recover(); r != nil {
			var message string =  fmt.Sprintf("server: start : simple: Errors dueing TLS Rest Server Start up, Details: %v", r)
			err = errors.New(message)
			if rs.logger != nil {
				rs.logger.Error(message)
			}
		} else {
			rs.tlsMode = false
		}
		if locked {
			rs.Unlock()
		}
	}()
	rs.Lock()
	locked = true
	if rs.server != nil {
		var mode string = "normal"
		if rs.tlsMode {
			mode = "TLS/TCP"
		}
		var message = fmt.Sprintf("server: start : simple: Server already started in %s mode!!", mode)
		if rs.logger != nil {
			rs.logger.Error(message)
		}
		return errors.New(fmt.Sprintf("server: start : simple: Server already started in %s mode!!", mode))
	}
	if rs.handlerFunc == nil {
		rs.server = &http.Server{
			Addr: fmt.Sprintf("%s:%v", hostOrIpAddress, port),
			TLSConfig: rs.config,
			Handler: rs,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context{
				ctx = context.WithValue(ctx, ncom.ContextRemoteAddress, c.RemoteAddr())
				sessionKey, err := uuid.NewV4()
				if err == nil {
					ctx = context.WithValue(ctx, ncom.ContextSessionKey, sessionKey)
				} else {
					if rs.logger != nil {
						rs.logger.Errorf("server: start : simple: Error retriving session id, Details: %s", err)
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
		rs.Unlock()
		locked = false
		err = rs.server.ListenAndServe()
		if err != nil && "http: Server closed" != err.Error() {
			rs.logger.Errorf("server: start : tls: Error: %s", err)
		}
	} else {
		service := fmt.Sprint("%s:%v",hostOrIpAddress, port)
		var list net.Listener
		list, err = tls.Listen("tcp", service, &tls.Config{})
		if err != nil {
			rs.logger.Fatalf("server: listen: Error: %s", err)
			if rs.listener != nil {
				(*rs.listener).Close()
			}
			panic("server: listen: " + err.Error())
		}
		rs.listener = &list
		rs.logger.Infof("server: listen: %v", service)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprintf("%v", r))
					rs.logger.Fatalf("TCP Server exit ...")
				}
				rs.logger.Info("TCP Server exit ...")
			}()
			for rs.IsRunning() {
				conn, errN := (*rs.listener).Accept()
				if errN != nil {
					rs.logger.Errorf("server: accept: %s", errN)
					continue
				}
				rs.logger.Debugf("server: accepted from %s", conn.RemoteAddr())
				tlscon, ok := conn.(*tls.Conn)
				if ok {
					rs.logger.Debug("ok=true")
					state := tlscon.ConnectionState()
					for _, v := range state.PeerCertificates {
						rs.logger.Info(x509.MarshalPKIXPublicKey(v.PublicKey))
					}
				}
				rs.conn = append(rs.conn, tlscon)
				go func(tlsconn *tls.Conn, server common.RestServer, restStruct *restServer) {
					restStruct.handlerFunc(tlsconn, rs)
					tlsconn.Close()
					defer func() {
						if r := recover(); r != nil {
							restStruct.logger.Errorf("Error closing connection, error: %v", r)
						}
						restStruct.Unlock()
					}()
					restStruct.Lock()
					conns := make([]*tls.Conn, 0)
					for _, c := range rs.conn {
						if c != tlsconn {
							conns = append(conns, c)
						}
					}
					restStruct.conn = conns
				}(tlscon, rs, rs)
			}
		}()
		rs.Unlock()
		locked = false
	}
	return err
}

func (rs *restServer) Stop() error {
	if rs.IsRunning() {
		if rs.server != nil {
			defer func() {
				rs.server = nil
			}()
			return rs.server.Close()
		} else if rs.listener != nil {
			defer func() {
				for _, c := range rs.conn {
					if c != nil {
						c.Close()
					}
				}
				rs.listener = nil
			}()
			return (*rs.listener).Close()
		}
	}
	return nil
}

func (rs *restServer) Shutdown() error {
	if rs.IsRunning() {
		if rs.server != nil {
			defer func() {
				rs.server = nil
			}()
			return rs.server.Shutdown(context.Background())
		} else {
			rs.listener = nil
			for _, c := range rs.conn {
				if c != nil {
					c.Close()
				}
			}
			return (*rs.listener).Close()
		}
	}
	return nil
}

func (rs *restServer) IsRunning() bool {
	return rs.server != nil || rs.listener != nil
}

func (rs *restServer) WaitFor() error{
	for rs.server != nil {
		time.Sleep(1 * time.Second)
	}
	return nil
}