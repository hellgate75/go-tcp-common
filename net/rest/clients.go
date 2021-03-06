package rest

import (
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/tls/client"
)

// Generate New TLS Client using CA Certificate file
func NewTLSClientWithCaCertificate(caCert string, ipAddress string, port string, logger log.Logger) common.RestClient {
	return client.NewWithCaCertificate(caCert, ipAddress, port, logger)
}

// Generate New TLS Client using Key and Certificate file
func NewTLSClientWithKeyCertificate(cert string, key string, ipAddress string, port string, logger log.Logger) common.RestClient {
	return client.NewWithCertificate(cert, key, ipAddress, port, logger)
}
