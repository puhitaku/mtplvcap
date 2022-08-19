package mtp

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type MJPEGResponseWriter struct {
	boundary string
	w        http.ResponseWriter
}

func NewMJPEGResponseWriter(w http.ResponseWriter) *MJPEGResponseWriter {
	boundary := randomBoundary()

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)

	return &MJPEGResponseWriter{
		boundary: boundary,
		w:        w,
	}
}

func (m *MJPEGResponseWriter) Write(jpeg []byte) error {
	f, ok := m.w.(http.Flusher)
	if !ok {
		return fmt.Errorf("HTTP buffer flushing is not implemented")
	}

	m.w.Write([]byte("Content-Type: image/jpeg\r\nContent-Length: " + strconv.Itoa(len(jpeg)) + "\r\n\r\n"))
	m.w.Write(jpeg)
	m.w.Write([]byte("\r\n--" + m.boundary + "\r\n"))

	f.Flush()

	return nil
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}
