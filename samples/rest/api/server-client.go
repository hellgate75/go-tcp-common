package main

import (
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-common/net/api/client"
	common2 "github.com/hellgate75/go-tcp-common/net/api/common"
	"github.com/hellgate75/go-tcp-common/net/api/server"
	"github.com/hellgate75/go-tcp-common/net/common"
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
	// Cosnumes (common.MimeType) - What this service consumes
	// Produces (common.MimeType) - What this service produces (MOCKED DEFAULT JSON)
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
	// We avoid state answer in case of success ...
	apiServer.AddApiAction("/hello", &helloAction{message: "{\"status\":\"OK\", \"message\": \"Hello there...\"}"},
		true, &method, &mimeType, &mimeType)
	go func(){
		err := apiServer.StartTLS("", port, &common2.TLSConfig{
			CaCertificate: "certs/ca.crt",
			CertFile: "certs/server.pem",
			KeyFile: "certs/server.key",
			UseInsecure: true,
		})
		if err != nil {
			serverLogger.Fatal("Unable to start Api Server in SSL/TLS security mode...")
			serverLogger.Fatalf("Error Details: %s", err)
			os.Exit(1)
		}
	}()
	time.Sleep(10 * time.Second)
	go func(){
		err := apiClient.ConnectTSL(ipAddress, port, &common2.TLSConfig{
			CaCertificate: "certs/ca.crt",
			CertFile: "certs/client.pem",
			KeyFile: "certs/client.key",
			UseInsecure: true,
		})
		if err != nil {
			serverLogger.Fatal("Unable to start Api Server in SSL/TLS security mode...")
			serverLogger.Fatalf("Error Details: %s", err)
			os.Exit(2)
		}
	}()
	time.Sleep(10 * time.Second)
	responseCode, answer, errCall :=  apiClient.GetApi(common.REST_PROTOCOL_HTTPS, "/hello", &method, &mimeType, &mimeType, nil, nil)
	clientLogger.Infof("Response Code: %v", responseCode)
	clientLogger.Infof("Error: %v", errCall)
	clientLogger.Infof("Message: %s", answer)
}
