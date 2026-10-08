package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
)

var catalogMagic = []byte("PFCA\x01")

// Options bounds catalog staging. InlineBytes is payload bytes; CatalogMiB
// bounds the encoded catalog; MaxObjects bounds its number of entries.
type Options struct {
	InlineBytes int
	CatalogMiB  int
	MaxObjects  int
}

func (s *Store) loadCatalog() error {
	s.inline = map[string][]byte{}
	s.catalogEstimate = 9
	file := filepath.Join(s.root, "catalog.pfca")
	if err := checkPath(file); err != nil {
		return readError(err)
	}
	info, err := os.Lstat(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return readError(err)
	}
	if !info.Mode().IsRegular() || info.Size() < 9 {
		return newError(Corrupt, "invalid object catalog file")
	}
	if info.Size() > s.catalogLimit {
		return newError(LimitExceeded, "object catalog exceeds capacity")
	}
	stream, err := os.Open(file)
	if err != nil {
		return readError(err)
	}
	defer stream.Close()
	data := make([]byte, int(info.Size()))
	if _, err = io.ReadFull(stream, data); err != nil {
		return readError(err)
	}
	var extra [1]byte
	if n, err := stream.Read(extra[:]); n != 0 || err != io.EOF {
		return newError(Corrupt, "catalog changed size while reading")
	}
	if bytes.Equal(data[:4], catalogMagic[:4]) && data[4] != catalogMagic[4] {
		return newError(UnsupportedFormat, "unsupported object catalog version")
	}
	if !bytes.Equal(data[:5], catalogMagic) {
		return newError(Corrupt, "invalid object catalog schema")
	}
	count := binary.BigEndian.Uint32(data[5:9])
	if uint64(count) > uint64(s.catalogCount) {
		return newError(LimitExceeded, "object catalog count exceeded")
	}
	cursor := 9
	for range count {
		if len(data)-cursor < 36 {
			return newError(Corrupt, "truncated catalog entry")
		}
		hash := hex.EncodeToString(data[cursor : cursor+32])
		size := uint64(binary.BigEndian.Uint32(data[cursor+32 : cursor+36]))
		cursor += 36
		if size < uint64(len(magic)+1) || size > MaxChunk+uint64(len(magic)+1) ||
			size > uint64(len(data)-cursor) {
			return newError(Corrupt, "invalid inline object capacity")
		}
		wire := data[cursor : cursor+int(size)]
		cursor += int(size)
		if _, found := s.inline[hash]; found {
			return newError(Corrupt, "duplicate catalog object")
		}
		if err = verifyWire(wire, hash); err != nil {
			return readError(err)
		}
		s.inline[hash] = wire
		s.catalogEstimate += 36 + int64(size)
	}
	if cursor != len(data) {
		return newError(Corrupt, "trailing catalog data")
	}
	return nil
}

func verifyWire(data []byte, hash string) error {
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != hash || len(data) < len(magic)+1 ||
		!bytes.Equal(data[:4], magic[:4]) {
		return newError(Corrupt, "object hash or format mismatch")
	}
	if data[4] != magic[4] {
		return newError(UnsupportedFormat, "unsupported content object version")
	}
	return nil
}

// A bounded binary catalog batches immutable objects without JSON/hex copies or many opens.
func (s *Store) stageInline(ref Ref, data []byte) bool {
	cost := int64(36 + len(data))
	if s.inlineBytes == 0 || len(data)-len(magic)-1 > s.inlineBytes ||
		len(s.inline) >= s.catalogCount || cost > s.catalogLimit-s.catalogEstimate {
		return false
	}
	s.inline[ref.Hash] = append([]byte(nil), data...)
	s.catalogEstimate += cost
	s.catalogDirty = true
	return true
}

// Publish objects before their version, preserving all existing hash-to-bytes entries.
func (s *Store) Flush() error {
	if !s.catalogDirty {
		return nil
	}
	data := make([]byte, 9, int(s.catalogEstimate))
	copy(data, catalogMagic)
	binary.BigEndian.PutUint32(data[5:9], uint32(len(s.inline)))
	hashes := make([]string, 0, len(s.inline))
	for hash := range s.inline {
		hashes = append(hashes, hash)
	}
	sort.Strings(hashes)
	for _, hash := range hashes {
		encoded, _ := hex.DecodeString(hash)
		data = append(data, encoded...)
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(s.inline[hash])))
		data = append(data, size[:]...)
		data = append(data, s.inline[hash]...)
	}
	if int64(len(data)) > s.catalogLimit {
		return newError(LimitExceeded, "object catalog exceeds capacity")
	}
	if err := checkPath(s.root); err != nil {
		return writeError(err)
	}
	if err := os.MkdirAll(s.root, 0700); err != nil {
		return writeError(err)
	}
	file := filepath.Join(s.root, "catalog.pfca")
	if err := checkPath(file); err != nil {
		return writeError(err)
	}
	pending, err := os.CreateTemp(s.root, ".pending-catalog-*.pfca")
	if err != nil {
		return writeError(err)
	}
	name := pending.Name()
	defer os.Remove(name)
	if _, err = pending.Write(data); err == nil {
		err = pending.Sync()
	}
	closeErr := pending.Close()
	if err != nil {
		return writeError(err)
	}
	if closeErr != nil {
		return writeError(closeErr)
	}
	if err = os.Rename(name, file); err != nil {
		return writeError(err)
	}
	s.catalogDirty = false
	return nil
}
