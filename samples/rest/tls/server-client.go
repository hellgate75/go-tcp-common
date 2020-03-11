package main

import (
	"fmt"
	"github.com/hellgate75/go-tcp-common/net/rest/tls/client"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/rest/common"
	"github.com/hellgate75/go-tcp-common/net/rest/tls/server"
	"net/http"
	"os"
	"time"
)

func main() {
	var port int32 = 40990
	var serverLogger log.Logger = log.NewLogger("test rest server", log.DEBUG)
	var clientLogger log.Logger = log.NewLogger("test rest client", log.DEBUG)
	caCert := "certs\\ca.crt"
	serverCert := "certs\\server.pem"
	serverKey := "certs\\server.key"
	server := server.NewCaCert(true, caCert, serverLogger)
	client := client.NewWithCaCertificate(caCert, "127.0.0.1", fmt.Sprintf("%v", port), clientLogger)
	mime := common.PLAIN_TEXT_MIME_TYPE
	methods := make([]common.RestMethod, 0)
	methods = append(methods, common.REST_METHOD_GET)
	if ok := server.AddPath("/hello",
		func(w http.ResponseWriter, req *http.Request, path string, accepts common.MimeType, produces common.MimeType) () {
			common.SubmitSuccess(w, "Hello!!")
		}, &mime, &mime, methods); !ok {
		serverLogger.Error("Unable to add server pattern!!")
		os.Exit(1)
	}
	time.Sleep(3 * time.Second)
	var errS error = nil
	var errC error = nil
	defer client.Close()
	defer server.Stop()
	go func() {
		errS = server.StartTLS("", port, serverCert, serverKey)
		if errS != nil {
			serverLogger.Errorf("Unable to start server: %s", errS)
			os.Exit(2)
		}
	}()
	time.Sleep(5 * time.Second)
	go func() {
		errC = client.Open()
		if errC != nil {
			serverLogger.Errorf("Unable to start client: %s", errC)
			os.Exit(3)
		}
	}()
	time.Sleep(10 * time.Second)
	//(int, []byte, error)
	responseCode, answer, errCall :=  client.Request(common.REST_PROTOCOL_HTTPS,"/hello", common.REST_METHOD_GET, &mime, nil, nil)
	clientLogger.Infof("Response Code: %v", responseCode)
	clientLogger.Infof("Error: %v", errCall)
	clientLogger.Infof("Message: %s", answer)
}
