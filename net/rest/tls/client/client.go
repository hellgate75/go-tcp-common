package client

import (
	"crypto/tls"
	"github.com/hellgate75/go-tcp-common/common"
	"github.com/hellgate75/go-tcp-common/log"
	"net/http"
)

type restClient struct {
	useInsecure     bool
	Cert            *common.CertificateKeyPair
	CaCert          string
	IpAddress       string
	Port            string
	client          *http.Client
	conn            *tls.Conn
	logger          log.Logger
}

func NewWithCertificate(cert string, key string, ipAddress string, port string, logger log.Logger) common.RestClient {
	return &restClient {
		Cert: &common.CertificateKeyPair{
			Cert: cert,
			Key: key,
		},
		IpAddress: ipAddress,
		Port: port,
		client: nil,
		logger: logger,
		conn: nil,
		CaCert: "",
		useInsecure: false,
	}
}

func NewWithCaCertificate(caCert string, ipAddress string, port string, logger log.Logger) common.RestClient {
	return &restClient {
		Cert: nil,
		IpAddress: ipAddress,
		Port: port,
		client: nil,
		logger: logger,
		conn: nil,
		CaCert: caCert,
		useInsecure: true,
	}
}
