package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	ccom "github.com/hellgate75/go-tcp-common/net/rest/common"
	"net/url"
)

func (rc *restClient) Request(protocol ccom.RestProtocol, path string, method ccom.RestMethod, accepts *ccom.MimeType, body *[]byte, values *url.Values) (int, []byte, error) {
	var html *http.Response = nil
	var err error = nil
	var status int = 0
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("RestClient.Connect, Error: %s", r))
			status = 500
		}
	}()
	requestUrl := fmt.Sprintf("%s://%s:%s%s", string(protocol), rc.IpAddress, rc.Port, path)
	if ccom.REST_METHOD_GET == method {
		html,err = rc.client.Get(requestUrl)
	} else if ccom.REST_METHOD_POST == method {
		html,err = rc.client.Post(requestUrl, string(*accepts), bytes.NewBuffer(*body))
	} else if ccom.REST_METHOD_POST_FORM == method {
		html,err = rc.client.PostForm(requestUrl, *values)
	} else if ccom.REST_METHOD_HEAD == method {
		html,err = rc.client.Head(requestUrl)
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
		rc.logger.Errorf("Status: %v", html.StatusCode)
		rc.logger.Errorf("Error reading body: %v", err)
		return status, []byte{}, err
	}
	rc.logger.Debugf("Status: %v", html.StatusCode)
	status = html.StatusCode
	return status, output, err
}

func (rc *restClient) Open() error {
	var err error = nil
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("RestClient.Connect, Error: %s", r))
		}
	}()
	if rc.client != nil {
		return errors.New("Client already cinnected!!")
	}
	
	var config *tls.Config = &tls.Config{}
	if "" != rc.CaCert {
		rc.logger.Debugf("client: using ca cert: <%s>", rc.CaCert)
		caCert, err := ioutil.ReadFile(rc.CaCert)
		if err != nil {
			rc.logger.Errorf("client: error using ca cert: <%s>, details: %s", rc.CaCert, err.Error())
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				rc.logger.Error("No certs appended, using system certificates only")
				rc.logger.Fatalf("client: loadkeys: %s", err.Error())
			} else {
				config.RootCAs= caCertPool
				config.InsecureSkipVerify = true
			}
		}
	}
	if rc.Cert != nil && "" != rc.Cert.Key &&  "" != rc.Cert.Cert {
		rc.logger.Debugf("client: using client key: <%s>, cert: <%s> ", rc.Cert.Key, rc.Cert.Cert)
		rc.logger.Debugf("client: using client key: <%s>, cert: <%s> ", rc.Cert.Key, rc.Cert.Cert)
		cert, err := tls.LoadX509KeyPair(rc.Cert.Cert, rc.Cert.Key)
		if err != nil {
			rc.logger.Errorf("client: Unable to load key : %s and certificate: %s", rc.Cert.Key, rc.Cert.Cert)
			rc.logger.Fatalf("client: loadkeys: %s", err.Error())
		} else {
			config.Certificates=[]tls.Certificate{cert}
		}
	}
	rc.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}
	service := fmt.Sprintf("%s:%s", rc.IpAddress, rc.Port)
	rc.logger.Debugf("Connecting to service: %s", service)
	conn, err := tls.Dial("tcp", service, config)
	if err != nil {
		rc.logger.Fatalf("client: dial: %s", err)
		return errors.New(fmt.Sprintf("client: dial: %s", err))
	}
	rc.conn = conn
	rc.logger.Debugf("client: connected to: %v", conn.RemoteAddr())
	state := conn.ConnectionState()
	rc.logger.Trace("Uaing certificates: ")
	for _, v := range state.PeerCertificates {
		bytes, errBts := x509.MarshalPKIXPublicKey(v.PublicKey)
		if errBts == nil {
			rc.logger.Trace("Public Key: ", string(bytes))
		} else {
			rc.logger.Trace("Public Key: Unavailable")
		}
		rc.logger.Trace(v.Subject)
	}
	rc.logger.Trace("client: handshake: ", state.HandshakeComplete)
	rc.logger.Trace("client: mutual: ", state.NegotiatedProtocolIsMutual)
	rc.logger.Debug("client: Connected!!")
	return err
}

