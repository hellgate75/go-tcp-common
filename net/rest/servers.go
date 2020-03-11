package rest

import (
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/tls/server"
)

func NewTLSRestServer(logger log.Logger) server.RestServer {
	return server.New(logger)
}

func NewCaCertTLSRestServer(allowInsecureConnections bool, caCert string, logger log.Logger) server.RestServer {
	return server.NewCaCert(allowInsecureConnections, caCert, logger)
}
