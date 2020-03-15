package main

import (
	"github.com/hellgate75/go-deploy/types/module"
	"github.com/hellgate75/go-tcp-common/io/streams"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/api/client"
	common2 "github.com/hellgate75/go-tcp-common/net/api/common"
	"github.com/hellgate75/go-tcp-common/net/api/server"
	"github.com/hellgate75/go-tcp-common/net/common"
	common3 "github.com/hellgate75/go-tcp-common/net/rest/common"
	"io"
	"net/http"
	"os"
	"time"
)

type helloAction struct {
	message string
}

func (ha *helloAction) Run(Args ...interface{}) error {
	// It receives:
	// req (*http.Request) - Server Handling Request
	// w (http.ResponseWriter) - Server Handling Output Writer
	// method (string) - web method
	// Cosnumes (types.MimeType) - What this service consumes
	// Produces (types.MimeType) - What this service produces (MOCKED DEFAULT JSON)
	common.SubmitSuccess(Args[1].(http.ResponseWriter), ha.message)
	return nil
}

func main() {
	serverLogger := log.NewLogger("api-server", log.DEBUG)
	clientLogger := log.NewLogger("api-client", log.DEBUG)
	apiServer := server.NewApiServer(serverLogger)
	apiClient := client.NewApiClient(clientLogger)
	ipAddress := "127.0.0.1"
	var port int64 = 40899
	method := common.REST_METHOD_GET
	mimeType := common.JSON_MIME_TYPE
	mimeType3 := common.PLAIN_TEXT_MIME_TYPE
	// We avoid state answer in case of success ...
	apiServer.AddApiAction("/hello", &helloAction{message: "{\"status\":\"OK\", \"message\": \"Hello there...\"}"},
		true, &method, &mimeType, &mimeType)
	reader, writer := io.Pipe()
	stream, error1 := streams.NewPipeStream(reader)
	if error1 != nil {
		serverLogger.Fatal("Unable to start create pipe stream...")
		serverLogger.Fatalf("Error Details: %s", error1)
		os.Exit(5)
	}
	apiServer.AddApiStream("/sessionIds", stream, &method, &mimeType3, &mimeType3)
	go func(){
		err := apiServer.StartTLS("", port, &common2.TLSConfig{
			CaCertificate: "certs/ca.crt",
			Certificates: []common3.CertificateKeyPair{
				common3.CertificateKeyPair{
					Cert: "certs/server.pem",
					Key: "certs/server.key",
				},
			},
			UseInsecure: true,
		})
		if err != nil {
			serverLogger.Fatal("Unable to start Api Server in SSL/TLS security mode...")
			serverLogger.Fatalf("Error Details: %s", err)
			os.Exit(1)
		}
	}()
	time.Sleep(10 * time.Second)
	defer func() {
		time.Sleep(5 * time.Second)
		apiClient.Close()
		apiServer.Stop()
	}()
	go func(){
		err := apiClient.ConnectTSL(ipAddress, port, &common2.TLSConfig{
			CaCertificate: "certs/ca.crt",
			Certificates: []common3.CertificateKeyPair{
				common3.CertificateKeyPair{
					Cert: "certs/client.pem",
					Key: "certs/client.key",
				},
			},
			UseInsecure: true,
		})
		if err != nil {
			serverLogger.Fatal("Unable to start Api Server in SSL/TLS security mode...")
			serverLogger.Fatalf("Error Details: %s", err)
			os.Exit(2)
		}
	}()
	go func() {
		for i := 0; i< 10; i++ {
			writer.Write([]byte(module.NewSessionId() + "\n"))
			time.Sleep(150 * time.Millisecond)
		}
		writer.Close()
	}()
	time.Sleep(10 * time.Second)
	responseCode, answer, errCall :=  apiClient.GetApi(common.REST_PROTOCOL_HTTPS, "/hello", &method, &mimeType, &mimeType, nil, nil)
	clientLogger.Infof("API pattern: /hello")
	clientLogger.Infof("Response Code: %v", responseCode)
	clientLogger.Infof("Error: %v", errCall)
	clientLogger.Infof("Message: %s", answer)
	responseCode, answer, errCall =  apiClient.GetApi(common.REST_PROTOCOL_HTTPS, "/sessionIds", &method, &mimeType3, &mimeType3, nil, nil)
	clientLogger.Infof("API pattern: /sessionIds")
	clientLogger.Infof("Response Code: %v", responseCode)
	clientLogger.Infof("Error: %v", errCall)
	clientLogger.Infof("Provided plain session id(s): \n%s", answer)

	reader.Close()
}
