package graft

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory"
)

// Magic bytes for .graft files: PHLO
var MagicBytes = []byte{0x50, 0x48, 0x4C, 0x4F}

// Version 1
const Version = 1

// Manifest describes the graft metadata
type Manifest struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	MemoryCount int       `json:"memory_count"`
	Tags        []string  `json:"tags"`
}

// Payload is the JSON content inside the gzip stream
type Payload struct {
	Manifest  Manifest          `json:"manifest"`
	Memories  []memory.Memory   `json:"memories"`
	Citations []memory.Citation `json:"citations,omitempty"`
}

// Package creates a .graft file from memories
func Package(manifest Manifest, memories []memory.Memory, citations []memory.Citation, outputPath string) error {
	// Create payload
	payload := Payload{
		Manifest:  manifest,
		Memories:  memories,
		Citations: citations,
	}

	// Create file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Write Magic Bytes
	if _, err := f.Write(MagicBytes); err != nil {
		return fmt.Errorf("failed to write magic bytes: %w", err)
	}

	// Write Version
	if err := binary.Write(f, binary.LittleEndian, uint8(Version)); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}

	// Create Gzip Writer
	gz := gzip.NewWriter(f)
	defer gz.Close()

	// Encode JSON to Gzip
	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}

	return nil
}

// Unpack reads a .graft file
func Unpack(inputPath string) (*Payload, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Check Magic Bytes
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %w", err)
	}
	for i := 0; i < 4; i++ {
		if magic[i] != MagicBytes[i] {
			return nil, fmt.Errorf("invalid file format: not a .graft file")
		}
	}

	// Check Version
	var version uint8
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if version != Version {
		return nil, fmt.Errorf("unsupported version: %d (expected %d)", version, Version)
	}

	// Read Gzip
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	// Decode JSON
	var payload Payload
	if err := json.NewDecoder(gz).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	return &payload, nil
}

// Inspect returns just the manifest without fully loading memories
// Note: Currently we have to decompress the stream, but we could optimize
// by putting manifest first in a separate block in v2.
func Inspect(inputPath string) (*Manifest, error) {
	payload, err := Unpack(inputPath)
	if err != nil {
		return nil, err
	}
	return &payload.Manifest, nil
}
