package rest

import (
	"github.com/hellgate75/go-tcp-common/common"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/tls/server"
)

func NewTLSRestServer(logger log.Logger) common.RestServer {
	return server.New(logger)
}

func NewCaCertTLSRestServer(allowInsecureConnections bool, caCert string, logger log.Logger) common.RestServer {
	return server.NewCaCert(allowInsecureConnections, caCert, logger)
}
