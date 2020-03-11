package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"github.com/satori/go.uuid"
	"net"
	"net/http"
	"time"
)

func (rs *restServer) AddPath(path string, callback common.RestCallback, accepts *common.MimeType, produces *common.MimeType, allowedMethods []common.RestMethod) bool {
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
					common.SubmitFaiure(w, http.StatusInternalServerError, fmt.Sprintf("Error: %v", r))
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
					common.SubmitFaiure(w, http.StatusMethodNotAllowed, message)
					return
				}
				rs.logger.Warnf("server: exec-path: Calling path: %s, func: %v", path, handlerStruct.Handler != nil)
				if handlerStruct.Handler != nil {
					(*handlerStruct.Handler)(w, req, path, *(*handlerStruct).Consumes, *(*handlerStruct).Consumes)
				} else {
					rs.logger.Warnf("server: exec-path: Unavailable Handler for path: %s", path)
				}
			} else {
				common.SubmitFaiure(w, http.StatusNotFound, "NOT_FOUND")
				return
			}
		}
		rs.HandleFunc(path, handlerFunc)
		rs.RUnlock()
		state = true
	}
	return state
}

func (rs *restServer) AddRootPath(callback common.RestCallback, accepts *common.MimeType, produces *common.MimeType, allowedMethods []common.RestMethod) bool {
	return rs.AddPath("/", callback, accepts, produces, allowedMethods)
}

func (rs *restServer) StartTLS(hostOrIpAddress string, port int32, cert string, key string) error {
	var err error = nil
	var locked bool = false
	defer func() {
		if r := recover(); r != nil {
			var message string =  fmt.Sprintf("server: start : tls: Errors dueing TLS Rest Server Start up, Details: %v", r)
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
	rs.server = &http.Server{
		Addr: fmt.Sprintf("%s:%v", hostOrIpAddress, port),
		TLSConfig: rs.config,
		Handler: rs,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context{
			ctx = context.WithValue(ctx, common.ContextRemoteAddress, c.RemoteAddr())
			sessionKey, err := uuid.NewV4()
			if err == nil {
				ctx = context.WithValue(ctx, common.ContextSessionKey, sessionKey)
			} else {
				if rs.logger != nil {
					rs.logger.Errorf("server: start : tls: Error retriving session id, Details: %s", err)
				}
			}
			ctx = context.WithValue(ctx, common.ContextKeyAuthtoken, common.GenerateSecureToken(64))
			return ctx
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		ReadTimeout: DEFAULT_READ_TIMEOUT,
		ReadHeaderTimeout: DEFAULT_HEADER_READ_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
		IdleTimeout: DEFAULT_IDLE_TIMEOUT,
	}
	err = rs.server.ListenAndServeTLS(cert, key)
	if err != nil {
		rs.logger.Errorf("")
	}
	rs.Unlock()
	locked = false
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
	rs.server = &http.Server{
		Addr: fmt.Sprintf("%s:%v", hostOrIpAddress, port),
		TLSConfig: rs.config,
		Handler: rs,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context{
			ctx = context.WithValue(ctx, common.ContextRemoteAddress, c.RemoteAddr())
			sessionKey, err := uuid.NewV4()
			if err == nil {
				ctx = context.WithValue(ctx, common.ContextSessionKey, sessionKey)
			} else {
				if rs.logger != nil {
					rs.logger.Errorf("server: start : simple: Error retriving session id, Details: %s", err)
				}
			}
			ctx = context.WithValue(ctx, common.ContextKeyAuthtoken, common.GenerateSecureToken(64))
			return ctx
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		ReadTimeout: DEFAULT_READ_TIMEOUT,
		ReadHeaderTimeout: DEFAULT_HEADER_READ_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
		IdleTimeout: DEFAULT_IDLE_TIMEOUT,
	}
	err = rs.server.ListenAndServe()
	rs.Unlock()
	locked = false
	return err
}

func (rs *restServer) Stop() error {
	if rs.IsRunning() {
		return rs.server.Close()
	}
	return nil
}

func (rs *restServer) Shutdown() error {
	if rs.IsRunning() {
		return rs.server.Shutdown(context.Background())
	}
	return nil
}

func (rs *restServer) IsRunning() bool {
	return rs.server != nil
}

func (rs *restServer) WaitFor() error{
	for rs.server != nil {
		time.Sleep(1 * time.Second)
	}
	return nil
}