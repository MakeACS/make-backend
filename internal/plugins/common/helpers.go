package common

import (
	"bytes"
	"log"
	"net/http"
)

type WriteAdapter struct {
	Headers http.Header
	Body    bytes.Buffer
	Code    int
}

func (w *WriteAdapter) Header() http.Header {
	return w.Headers
}

func (w *WriteAdapter) Write(b []byte) (int, error) {
	log.Println("writing to adapter ", len(b), "bytes")
	return w.Body.Write(b)

}

func (w *WriteAdapter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func (w *WriteAdapter) IntoParts() (int, http.Header, []byte) {
	return w.Code, w.Headers, w.Body.Bytes()
}
