package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestObjectFailureCodesDistinguishCapacityCorruptionAndAbsence(t *testing.T) {
	store := fixtureStore(t)
	ref, err := store.PutText("原文🌿")
	if err != nil {
		t.Fatal(err)
	}
	assertCode := func(err error, expected Kind) {
		t.Helper()
		if code, found := kindOf(err); !found || code != expected {
			t.Fatal("wrong object failure code", code, expected, err)
		}
	}
	_, err = store.Session(1, 10).Text(ref, 1024)
	assertCode(err, LimitExceeded)
	file, err := store.location(ref.Hash)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 1
	if err = os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	_, err = store.Session(1024, 10).Text(ref, 1024)
	assertCode(err, Corrupt)
	if err = os.Remove(file); err != nil {
		t.Fatal(err)
	}
	_, err = store.Session(1024, 10).Text(ref, 1024)
	assertCode(err, Missing)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("missing dependency no longer unwraps")
	}
	ref, err = store.put(9, []byte("future"), 6)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Session(1024, 10).Text(ref, 1024)
	assertCode(err, UnsupportedFormat)
}

func TestFutureWireVersionIsUnsupportedRatherThanCorrupt(t *testing.T) {
	store := fixtureStore(t)
	wire := []byte{'P', 'F', 'C', 'O', 2, 0, 'a'}
	sum := sha256.Sum256(wire)
	ref := Ref{Hash: hex.EncodeToString(sum[:]), Bytes: 1}
	file, err := store.location(ref.Hash)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, wire, 0600); err != nil {
		t.Fatal(err)
	}
	_, err = store.Session(1024, 10).Text(ref, 1024)
	if code, _ := kindOf(err); code != UnsupportedFormat {
		t.Fatal("valid future wire hash misclassified", code, err)
	}
}
