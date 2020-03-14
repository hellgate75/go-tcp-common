package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-common/log"
	common2 "github.com/hellgate75/go-tcp-common/net/api/common"
	"github.com/hellgate75/go-tcp-common/net/common"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type apiClient struct{
	sync.Mutex
	logger 			log.Logger
	IpAddress       string
	Port            int64
	client          *http.Client
	connTls         *tls.Conn
}

func (cli *apiClient) Connect(ipAddress string, port int64) error {
	cli.IpAddress = ipAddress
	cli.Port = port
	var err error = nil
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("ApiClient.Connect, Error: %s", r))
		}
	}()
	if cli.client != nil {
		return errors.New("Client already cinnected!!")
	}
	var config *tls.Config = &tls.Config{}
	cli.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}
	service := fmt.Sprintf("%s:%v", cli.IpAddress, cli.Port)
	cli.logger.Debugf("Connecting to service: %s", service)
	conn, err := tls.Dial("tcp", service, config)
	if err != nil {
		cli.logger.Fatalf("client: dial: %s", err)
		return errors.New(fmt.Sprintf("client: dial: %s", err))
	}
	cli.connTls = conn
	return err
}
func (cli *apiClient) ConnectTSL(ipAddress string, port int64, baseConfig *common2.TLSConfig) error {
	cli.IpAddress = ipAddress
	cli.Port = port
	var err error = nil
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("ApiClient.Connect, Error: %s", r))
		}
	}()
	if cli.client != nil {
		return errors.New("Client already cinnected!!")
	}

	var config *tls.Config = &tls.Config{}
	if "" != baseConfig.CaCertificate {
		cli.logger.Debugf("client: using ca cert: <%s>", baseConfig.CaCertificate)
		caCert, err := ioutil.ReadFile(baseConfig.CaCertificate)
		if err != nil {
			cli.logger.Errorf("client: error using ca cert: <%s>, details: %s", baseConfig.CaCertificate, err.Error())
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				cli.logger.Error("No certs appended, using system certificates only")
				cli.logger.Fatalf("client: loadkeys: %s", err.Error())
			} else {
				config.RootCAs= caCertPool
				config.InsecureSkipVerify = true
			}
		}
	}
	if "" != baseConfig.KeyFile &&  "" != baseConfig.CertFile {
		cli.logger.Debugf("client: using client key: <%s>, cert: <%s> ", baseConfig.KeyFile, baseConfig.CertFile)
		cli.logger.Debugf("client: using client key: <%s>, cert: <%s> ", baseConfig.KeyFile, baseConfig.CertFile)
		cert, err := tls.LoadX509KeyPair(baseConfig.CertFile, baseConfig.KeyFile)
		if err != nil {
			cli.logger.Errorf("client: Unable to load key : %s and certificate: %s", baseConfig.KeyFile, baseConfig.CertFile)
			cli.logger.Fatalf("client: loadkeys: %s", err.Error())
		} else {
			config.Certificates=[]tls.Certificate{cert}
		}
	}
	cli.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}
	service := fmt.Sprintf("%s:%v", cli.IpAddress, cli.Port)
	cli.logger.Debugf("Connecting to service: %s", service)
	conn, err := tls.Dial("tcp", service, config)
	if err != nil {
		cli.logger.Fatalf("client: dial: %s", err)
		return errors.New(fmt.Sprintf("client: dial: %s", err))
	}
	cli.connTls = conn
	cli.logger.Debugf("client: connected to: %v", conn.RemoteAddr())
	state := conn.ConnectionState()
	cli.logger.Trace("Uaing certificates: ")
	for _, v := range state.PeerCertificates {
		bytes, errBts := x509.MarshalPKIXPublicKey(v.PublicKey)
		if errBts == nil {
			cli.logger.Trace("Public Key: ", string(bytes))
		} else {
			cli.logger.Trace("Public Key: Unavailable")
		}
		cli.logger.Trace(v.Subject)
	}
	cli.logger.Trace("client: handshake: ", state.HandshakeComplete)
	cli.logger.Trace("client: mutual: ", state.NegotiatedProtocolIsMutual)
	cli.logger.Debug("client: Connected!!")
	return err
}

func (cli *apiClient) Close() error {
	if cli.client != nil {
		cli.client.CloseIdleConnections()
		cli.client = nil
		if cli.connTls != nil {
			err := cli.connTls.Close()
			cli.connTls = nil
			return err
		}
		return nil
	}
	return nil
}
func (cli *apiClient) GetApi(protocol common.RestProtocol, path string, method *common.RestMethod, produces *common.MimeType, consumes *common.MimeType, body *[]byte, values *url.Values) (int, []byte, error) {
	var html *http.Response = nil
	var err error = nil
	var status int = 0
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("RestClient.Connect, Error: %s", r))
			status = 500
		}
	}()
	requestUrl := fmt.Sprintf("%s://%s:%v%s", string(protocol), cli.IpAddress, cli.Port, path)
	if common.REST_METHOD_GET == *method {
		html,err = cli.client.Get(requestUrl)
	} else if common.REST_METHOD_POST == *method {
		html,err = cli.client.Post(requestUrl, string(*consumes), bytes.NewBuffer(*body))
	} else if common.REST_METHOD_POST_FORM == *method {
		html,err = cli.client.PostForm(requestUrl, *values)
	} else if common.REST_METHOD_HEAD == *method {
		html,err = cli.client.Head(requestUrl)
	} else {
		return status, []byte{}, errors.New(fmt.Sprintf("Unavailable Method: %s!!", method))
	}
	if err!=nil {
		return status, []byte{}, err
	}
	if html.StatusCode != 200 {
		return html.StatusCode, []byte{}, errors.New(fmt.Sprintf("Status Code: %v, Message: %s", html.StatusCode, html.Status))
	}
	defer html.Body.Close()
	output, err := ioutil.ReadAll(html.Body)
	if err != nil {
		cli.logger.Errorf("Status: %v", html.StatusCode)
		cli.logger.Errorf("Error reading body: %v", err)
		return status, []byte{}, err
	}
	cli.logger.Debugf("Status: %v", html.StatusCode)
	status = html.StatusCode
	return status, output, err
	return 0, []byte{}, nil
}

func NewApiClient(logger log.Logger) common2.APIClient {
	return &apiClient{
		logger: logger,
		client: nil,
	}
}