package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// StoredFile is a dependency validated by a completed session, with encoded byte length.
// It is distinct from Ref, whose length describes expanded logical text.
type StoredFile struct {
	Hash  string
	Bytes int64
}

// VerifyFiles rechecks every byte in an already verified dependency closure.
// Callers must match the exact original/version metadata before reusing that closure.
func (s *Store) VerifyFiles(files []StoredFile, maxBytes int64, maxObjects int) error {
	if len(files) > maxObjects {
		return newError(LimitExceeded, "object count exceeded")
	}
	checked := map[string]bool{}
	seen := map[string]bool{}
	buffer := make([]byte, 32*1024)
	// Catalog buffering has its own bound. Only reachable bytes consume history capacity.
	var consumed int64
	for _, file := range files {
		if seen[file.Hash] || file.Bytes < int64(len(magic)+1) {
			return newError(Corrupt, "object dependency size or identity invalid")
		}
		if file.Bytes > maxBytes-consumed {
			return newError(LimitExceeded, "object history capacity exceeded")
		}
		seen[file.Hash] = true
		if wire, found := s.inline[file.Hash]; found {
			if int64(len(wire)) != file.Bytes {
				return newError(Corrupt, "inline object changed size")
			}
			if err := verifyWire(wire, file.Hash); err != nil {
				return readError(err)
			}
			consumed += file.Bytes
			continue
		}
		stream, info, err := s.openObject(file.Hash, checked)
		if err != nil {
			return readError(err)
		}
		if info.Size() != file.Bytes {
			stream.Close()
			return newError(Corrupt, "object changed size")
		}
		hash := sha256.New()
		count, err := io.CopyBuffer(hash, io.LimitReader(stream, file.Bytes+1), buffer)
		closeErr := stream.Close()
		if err != nil {
			return readError(err)
		}
		if closeErr != nil {
			return readError(closeErr)
		}
		if count != file.Bytes || hex.EncodeToString(hash.Sum(nil)) != file.Hash {
			return newError(Corrupt, "object hash mismatch")
		}
		consumed += count
	}
	return nil
}

func (s *Session) Files() []StoredFile {
	files := make([]StoredFile, 0, len(s.objects))
	for _, hash := range s.Hashes() {
		node := s.objects[hash]
		files = append(files, StoredFile{Hash: hash, Bytes: int64(len(magic) + 1 + len(node.payload))})
	}
	return files
}
