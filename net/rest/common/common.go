package common

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type MimeType string

const(
	JSON_MIME_TYPE MimeType =           "application/json"
	XML_MIME_TYPE MimeType =            "application/xml"
	YAML_MIME_TYPE MimeType =           "text/yaml"
	ZIP_ARCHIVE_MIME_TYPE MimeType =    "application/zip"
	BINARY_DATA_MIME_TYPE MimeType =    "application/octet-stream"
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
