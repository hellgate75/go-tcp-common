package rest

import (
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/tls/server"
)

// Generate New TLS Server
func NewTLSRestServer(logger log.Logger) common.RestServer {
	return server.New(logger)
}

// Generate New TLS Server using CA Certificate file
func NewTLSRestServerHandler(handler server.TLSHandleFunc, logger log.Logger) common.RestServer {
	return server.NewHandleFunc(handler, logger)
}
