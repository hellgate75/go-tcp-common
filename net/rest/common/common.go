package common

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Mime Type
type MimeType string

// Rest Web Method  Type
type RestMethod string

// Rest Protocol Type
type RestProtocol string

const(
	//Plain Text format Mime Type
	PLAIN_TEXT_MIME_TYPE MimeType =     "text/plain"
	//Json format Mime Type
	JSON_MIME_TYPE MimeType =           "application/json"
	//Xml format Mime Type
	XML_MIME_TYPE MimeType =            "application/xml"
	//Yaml format Mime Type
	YAML_MIME_TYPE MimeType =           "text/yaml"
	//Zip Archive format Mime Type
	ZIP_ARCHIVE_MIME_TYPE MimeType =    "application/zip"
	//Binary Data format Mime Type
	BINARY_DATA_MIME_TYPE MimeType =    "application/octet-stream"
	
	//Rest Web Method: GET
	REST_METHOD_GET RestMethod =        http.MethodGet
	//Rest Web Method: POST
	REST_METHOD_POST RestMethod =       http.MethodPost
	//Rest Web Method: POST (FORM)
	REST_METHOD_POST_FORM RestMethod =  http.MethodPost+"_FORM"
	//Rest Web Method: HEAD
	REST_METHOD_HEAD RestMethod =       http.MethodHead
	
	//Rest Web Protocol: HTTP
	REST_PROTOCOL_HTTP RestProtocol =   "http"
	//Rest Web Protocol: HTTPS
	REST_PROTOCOL_HTTPS RestProtocol =  "https"
	//Rest Web Protocol: WS
	REST_PROTOCOL_WS RestProtocol =     "ws"
)
// Context Key Type
type ContextKey string

func (c ContextKey) String() string {
	return "context key " + string(c)
}

var (
	// Session Context Key
	ContextSessionKey = ContextKey("session-key")
	// Session Context Auth Token
	ContextKeyAuthtoken = ContextKey("auth-token")
	// Session Context Remote Address
	ContextRemoteAddress   = ContextKey("remote-address")
)

// Generate a Security Token of a given length
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// Submit Success Data To the client
func SubmitSuccess(w http.ResponseWriter, message string) {
	w.WriteHeader(200)
	w.Write([]byte(message))
}

// Submit Failure Data To the client
func SubmitFaiure(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
