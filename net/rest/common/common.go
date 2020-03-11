package common

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type MimeType string
type RestMethod string
type RestProtocol string

const(
	PLAIN_TEXT_MIME_TYPE MimeType =     "text/plain"
	JSON_MIME_TYPE MimeType =           "application/json"
	XML_MIME_TYPE MimeType =            "application/xml"
	YAML_MIME_TYPE MimeType =           "text/yaml"
	ZIP_ARCHIVE_MIME_TYPE MimeType =    "application/zip"
	BINARY_DATA_MIME_TYPE MimeType =    "application/octet-stream"
	
	REST_METHOD_GET RestMethod =        http.MethodGet
	REST_METHOD_POST RestMethod =       http.MethodPost
	REST_METHOD_POST_FORM RestMethod =  http.MethodPost+"_FORM"
	REST_METHOD_HEAD RestMethod =       http.MethodHead
	
	REST_PROTOCOL_HTTP RestProtocol =   "http"
	REST_PROTOCOL_HTTPS RestProtocol =  "https"
	REST_PROTOCOL_WS RestProtocol =     "ws"
)
type ContextKey string
func (c ContextKey) String() string {
	return "context key " + string(c)
}
var (
	ContextSessionKey = ContextKey("session-key")
	ContextKeyAuthtoken = ContextKey("auth-token")
	ContextRemoteAddress   = ContextKey("remote-address")
)
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func SubmitSuccess(w http.ResponseWriter, message string) {
	w.WriteHeader(200)
	w.Write([]byte(message))
}

func SubmitFaiure(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
