package telemetry

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"google.golang.org/protobuf/proto"

	gimpelv1 "gimpel/api/go/v1"
)

type Buffer struct {
	path     string
	maxBytes int64

	mu       sync.Mutex
	file     *os.File
	position int64
}

func NewBuffer(path string, maxBytes int64) (*Buffer, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating buffer dir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening buffer file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("statting buffer file: %w", err)
	}

	return &Buffer{
		path:     path,
		maxBytes: maxBytes,
		file:     file,
		position: info.Size(),
	}, nil
}

func (b *Buffer) Push(event *gimpelv1.Event) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.position+int64(len(data))+4 > b.maxBytes {
		return fmt.Errorf("buffer full")
	}

	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(data)))

	if _, err := b.file.Seek(b.position, io.SeekStart); err != nil {
		return fmt.Errorf("seeking: %w", err)
	}

	if _, err := b.file.Write(sizeBytes); err != nil {
		return fmt.Errorf("writing size: %w", err)
	}

	if _, err := b.file.Write(data); err != nil {
		return fmt.Errorf("writing data: %w", err)
	}

	b.position += int64(4 + len(data))
	return nil
}

func (b *Buffer) Pop(count int) ([]*gimpelv1.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.position == 0 {
		return nil, nil
	}

	if _, err := b.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking to start: %w", err)
	}

	events := make([]*gimpelv1.Event, 0, count)
	readPos := int64(0)

	for len(events) < count && readPos < b.position {
		sizeBytes := make([]byte, 4)
		if _, err := io.ReadFull(b.file, sizeBytes); err != nil {
			break
		}
		size := binary.BigEndian.Uint32(sizeBytes)
		readPos += 4

		data := make([]byte, size)
		if _, err := io.ReadFull(b.file, data); err != nil {
			break
		}
		readPos += int64(size)

		var event gimpelv1.Event
		if err := proto.Unmarshal(data, &event); err != nil {
			continue
		}
		events = append(events, &event)
	}

	remaining := b.position - readPos
	if remaining > 0 {
		remainingData := make([]byte, remaining)
		if _, err := io.ReadFull(b.file, remainingData); err == nil {
			b.file.Seek(0, io.SeekStart)
			b.file.Write(remainingData)
			b.file.Truncate(remaining)
		}
	} else {
		b.file.Truncate(0)
	}

	b.position = remaining

	return events, nil
}

func (b *Buffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.file != nil {
		return b.file.Close()
	}
	return nil
}
