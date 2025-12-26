//go:build cgo
// +build cgo

package vax

/*
#cgo CFLAGS: -I${SRCDIR}/../../../c/include
#cgo LDFLAGS: -L${SRCDIR}/../../../c/build -lvax -lcrypto -lssl
#include <vax.h>
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

// Error codes matching C implementation
var (
	ErrInvalidCounter  = errors.New("invalid counter")
	ErrInvalidPrevSAI  = errors.New("invalid prevSAI")
	ErrSAIMismatch     = errors.New("SAI mismatch")
	ErrOutOfMemory     = errors.New("out of memory")
	ErrInvalidInput    = errors.New("invalid input")
	ErrCounterOverflow = errors.New("counter overflow")
)

// Constants
const (
	SAISize         = 32
	GISize          = 32
	KChainSize      = 32
	GenesisSaltSize = 16
)

// ComputeGI computes gi_n = HMAC_SHA256(k_chain, "VAX-GI" || counter)
func ComputeGI(kChain []byte, counter uint16) ([]byte, error) {
	if len(kChain) != KChainSize {
		return nil, ErrInvalidInput
	}

	var gi [GISize]C.uint8_t

	result := C.vax_compute_gi(
		(*C.uint8_t)(unsafe.Pointer(&kChain[0])),
		C.uint16_t(counter),
		&gi[0],
	)

	if result != C.VAX_OK {
		return nil, mapCError(result)
	}

	return C.GoBytes(unsafe.Pointer(&gi[0]), GISize), nil
}

// ComputeSAI computes SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)
func ComputeSAI(prevSAI, saeBytes, gi []byte) ([]byte, error) {
	if len(prevSAI) != SAISize {
		return nil, ErrInvalidInput
	}
	if len(gi) != GISize {
		return nil, ErrInvalidInput
	}
	if len(saeBytes) == 0 {
		return nil, ErrInvalidInput
	}

	var sai [SAISize]C.uint8_t

	result := C.vax_compute_sai(
		(*C.uint8_t)(unsafe.Pointer(&prevSAI[0])),
		(*C.uint8_t)(unsafe.Pointer(&saeBytes[0])),
		C.size_t(len(saeBytes)),
		(*C.uint8_t)(unsafe.Pointer(&gi[0])),
		&sai[0],
	)

	if result != C.VAX_OK {
		return nil, mapCError(result)
	}

	return C.GoBytes(unsafe.Pointer(&sai[0]), SAISize), nil
}

// ComputeGenesisSAI computes genesis SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
func ComputeGenesisSAI(actorID string, genesisSalt []byte) ([]byte, error) {
	if len(genesisSalt) != GenesisSaltSize {
		return nil, ErrInvalidInput
	}

	cActorID := C.CString(actorID)
	defer C.free(unsafe.Pointer(cActorID))

	var sai [SAISize]C.uint8_t

	result := C.vax_compute_genesis_sai(
		cActorID,
		(*C.uint8_t)(unsafe.Pointer(&genesisSalt[0])),
		&sai[0],
	)

	if result != C.VAX_OK {
		return nil, mapCError(result)
	}

	return C.GoBytes(unsafe.Pointer(&sai[0]), SAISize), nil
}

// VerifyAction verifies an action submission (crypto only, no JSON validation)
func VerifyAction(
	kChain []byte,
	expectedCounter uint16,
	expectedPrevSAI []byte,
	counter uint16,
	prevSAI []byte,
	saeBytes []byte,
	sai []byte,
) error {
	if len(kChain) != KChainSize {
		return ErrInvalidInput
	}
	if len(expectedPrevSAI) != SAISize {
		return ErrInvalidInput
	}
	if len(prevSAI) != SAISize {
		return ErrInvalidInput
	}
	if len(sai) != SAISize {
		return ErrInvalidInput
	}
	if len(saeBytes) == 0 {
		return ErrInvalidInput
	}

	result := C.vax_verify_action(
		(*C.uint8_t)(unsafe.Pointer(&kChain[0])),
		C.uint16_t(expectedCounter),
		(*C.uint8_t)(unsafe.Pointer(&expectedPrevSAI[0])),
		C.uint16_t(counter),
		(*C.uint8_t)(unsafe.Pointer(&prevSAI[0])),
		(*C.uint8_t)(unsafe.Pointer(&saeBytes[0])),
		C.size_t(len(saeBytes)),
		(*C.uint8_t)(unsafe.Pointer(&sai[0])),
	)

	if result != C.VAX_OK {
		return mapCError(result)
	}

	return nil
}

// mapCError converts C error codes to Go errors
func mapCError(result C.vax_result_t) error {
	switch result {
	case C.VAX_OK:
		return nil
	case C.VAX_ERR_INVALID_COUNTER:
		return ErrInvalidCounter
	case C.VAX_ERR_INVALID_PREV_SAI:
		return ErrInvalidPrevSAI
	case C.VAX_ERR_SAI_MISMATCH:
		return ErrSAIMismatch
	case C.VAX_ERR_OUT_OF_MEMORY:
		return ErrOutOfMemory
	case C.VAX_ERR_INVALID_INPUT:
		return ErrInvalidInput
	case C.VAX_ERR_COUNTER_OVERFLOW:
		return ErrCounterOverflow
	default:
		return errors.New("unknown error")
	}
}
