package client

import (
	"crypto/tls"
	"github.com/hellgate75/go-tcp-common/common"
	"github.com/hellgate75/go-tcp-common/log"
	ccom "github.com/hellgate75/go-tcp-common/net/rest/common"
	"net/http"
)

type RestClient interface{
	Open() error
	Request(path string, method string, contentType *ccom.MimeType, body *[]byte) (int, []byte, error)
}

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

func NewWithCertificate(cert string, key string, ipAddress string, port string, logger log.Logger) RestClient {
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

func NewWithCaCertificate(caCert string, ipAddress string, port string, logger log.Logger) RestClient {
	return &restClient {
		Cert: nil,
		IpAddress: ipAddress,
		Port: port,
		client: nil,
		logger: logger,
		conn: nil,
		CaCert: caCert,
		useInsecure: false,
	}
}
