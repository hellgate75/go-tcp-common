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
func NewCaCertTLSRestServer(allowInsecureConnections bool, caCert string, logger log.Logger) common.RestServer {
	return server.NewCaCert(allowInsecureConnections, caCert, logger)
}
