package mtp

import (
	"context"
	"net/http"
	"time"
)

type SnapshotResponseWriter struct {
	Context context.Context
	cancel  context.CancelFunc
	w       http.ResponseWriter
}

func NewSnapshotResponseWriter(w http.ResponseWriter) *SnapshotResponseWriter {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, cancel = context.WithCancel(ctx)

	return &SnapshotResponseWriter{
		Context: ctx,
		cancel:  cancel,
		w:       w,
	}
}

func (s *SnapshotResponseWriter) Write(jpeg []byte) error {
	s.w.Header().Set("Content-Type", "image/jpeg")
	s.w.Write(jpeg)
	s.cancel()

	return nil
}
