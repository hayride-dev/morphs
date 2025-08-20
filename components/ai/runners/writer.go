package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hayride-dev/bindings/go/hayride/types"
)

type MessageWriter struct {
	writerType types.WriterType
	w          io.Writer
}

func NewMessageWriter(writerType types.WriterType, w io.Writer) *MessageWriter {
	return &MessageWriter{
		writerType: writerType,
		w:          w,
	}
}

func (mw *MessageWriter) WriteMessage(message types.Message) error {
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	switch mw.writerType {
	case types.WriterTypeSse:
		if _, err := mw.w.Write([]byte("data: " + string(b) + "\n\n")); err != nil {
			return fmt.Errorf("failed to write SSE message: %w", err)
		}
		return nil
	case types.WriterTypeRaw:
		if _, err := mw.w.Write(b); err != nil {
			return fmt.Errorf("failed to write raw message: %w", err)
		}
		return nil
	}

	return fmt.Errorf("unknown writer type: %s", mw.writerType)
}
