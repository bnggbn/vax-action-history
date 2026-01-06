#ifndef VAX_H
#define VAX_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/*============================================================================
 * Constants
 *===========================================================================*/

#define VAX_SAI_SIZE       32
#define VAX_GI_SIZE        32
#define VAX_K_CHAIN_SIZE   32
#define VAX_GENESIS_SALT_SIZE 16

/*============================================================================
 * Error Codes
 *===========================================================================*/

typedef enum {
    VAX_OK = 0,
    VAX_ERR_INVALID_COUNTER,
    VAX_ERR_INVALID_PREV_SAI,
    VAX_ERR_SAI_MISMATCH,
    VAX_ERR_OUT_OF_MEMORY,
    VAX_ERR_INVALID_INPUT,
    VAX_ERR_COUNTER_OVERFLOW,
} vax_result_t;

/*============================================================================
 * Core Cryptographic Primitives
 *===========================================================================*/

/**
 * Compute SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)
 * gi is generated internally using random bytes.
 */
vax_result_t vax_compute_sai(
    const uint8_t prev_sai[VAX_SAI_SIZE],
    const uint8_t* sae_bytes,
    size_t sae_len,
    uint8_t out_sai[VAX_SAI_SIZE]
);

/**
 * Compute genesis SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
 */
vax_result_t vax_compute_genesis_sai(
    const char* actor_id,
    const uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE],
    uint8_t out_sai[VAX_SAI_SIZE]
);

/*============================================================================
 * Verification (Crypto Only)
 *===========================================================================*/

/**
 * Verify an action submission (crypto only, no JSON validation)
 *
 * This performs:
 * - Verify prevSAI matches expected
 *
 * NOTE: SAE canonicalization must be verified by the caller
 *       (e.g., Go server or JS frontend)
 */
vax_result_t vax_verify_action(
    const uint8_t expected_prev_sai[VAX_SAI_SIZE],
    const uint8_t prev_sai[VAX_SAI_SIZE]
);

/*============================================================================
 * Utility Functions
 *===========================================================================*/

/**
 * Convert error code to human-readable string
 */
const char* vax_error_string(vax_result_t result);

#ifdef __cplusplus
}
#endif

#endif /* VAX_H */
