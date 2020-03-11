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

func (rs *restServer) AddPath(path string, callback RestCallback, accepts *common.MimeType, produces *common.MimeType) bool {
	var state bool = false
	defer func() {
		if r := recover(); r != nil {
			if rs.logger != nil {
				rs.logger.Errorf("Errors dueing TLS Rest Server Start up, Details: %v", r)
			}
			rs.RUnlock()
			state = false
		}
	}()
	if _, ok := rs.paths[path]; !ok {
		rs.RLock()
		rs.paths[path] = &HandlerStruct{
				Handler: &callback,
				Consumes: accepts,
				Produces: produces,
				Path: path,
		}
		handlerFunc := func(w http.ResponseWriter, req *http.Request)() {
			defer func(){
				if r := recover(); r != nil {
					common.SubmitFaiure(w, 500, fmt.Sprintf("Error: %v", r))
				}
			}()
			var path string = req.URL.Path
			fmt.Printf("Path: %s", path)
			if handlerStruct, ok := rs.paths[path]; ok {
				(*handlerStruct.Handler)(w, req, path, *(*handlerStruct).Consumes, *(*handlerStruct).Consumes)
			} else {
				common.SubmitFaiure(w, 404, "NOT_FOUND")
			}
		}
		rs.HandleFunc(path, handlerFunc)
		rs.RUnlock()
		state = true
	}
	return state
}

func (rs *restServer) AddRootPath(callback RestCallback, accepts *common.MimeType, produces *common.MimeType) bool {
	return rs.AddPath("/", callback, accepts, produces)
}

func (rs *restServer) StartTLS(hostOrIpAddress string, port int32, cert string, key string) error {
	var err error = nil
	var locked bool = false
	defer func() {
		if r := recover(); r != nil {
			var message string =  fmt.Sprintf("Errors dueing TLS Rest Server Start up, Details: %v", r)
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
		var message = fmt.Sprintf("Server already started in %s mode!!", mode)
		if rs.logger != nil {
			rs.logger.Error(message)
		}
		return errors.New(fmt.Sprintf("Server already started in %s mode!!", mode))
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
					rs.logger.Errorf("Error retriving session id, Details: %s", err)
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
	rs.Unlock()
	locked = false
	return err
}

func (rs *restServer) Start(hostOrIpAddress string, port int32) error {
	var err error = nil
	var locked bool = false
	defer func() {
		if r := recover(); r != nil {
			var message string =  fmt.Sprintf("Errors dueing TLS Rest Server Start up, Details: %v", r)
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
		var message = fmt.Sprintf("Server already started in %s mode!!", mode)
		if rs.logger != nil {
			rs.logger.Error(message)
		}
		return errors.New(fmt.Sprintf("Server already started in %s mode!!", mode))
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
					rs.logger.Errorf("Error retriving session id, Details: %s", err)
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