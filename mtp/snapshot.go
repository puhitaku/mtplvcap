package mtp

import (
	"net/http"
)

type SnapshotResponseWriter struct {
	w http.ResponseWriter
}

func NewSnapshotResponseWriter(w http.ResponseWriter) *SnapshotResponseWriter {
	return &SnapshotResponseWriter{
		w: w,
	}
}

func (s *SnapshotResponseWriter) Write(jpeg []byte) error {
	s.w.Header().Set("Content-Type", "image/jpeg")
	s.w.Write(jpeg)

	return nil
}
