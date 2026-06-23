package middleware

import "net/http"

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
	rr.ResponseWriter.WriteHeader(status)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.bytes += int64(n)
	return n, err
}
