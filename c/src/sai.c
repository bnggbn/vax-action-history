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
    VAX_ERR_INVALID_CANONICALIZATION,
    VAX_ERR_SAI_MISMATCH,
    VAX_ERR_GI_MISMATCH,
    VAX_ERR_OUT_OF_MEMORY,
    VAX_ERR_INVALID_INPUT,
    VAX_ERR_ACTOR_MISMATCH,
} vax_result_t;

/*============================================================================
 * Low-Level API (純函數，無狀態)
 *===========================================================================*/

/**
 * Compute gi_n = HMAC_SHA256(k_chain, "VAX-GI" || counter)
 * 
 * @param k_chain   32-byte session secret
 * @param counter   Actor-scoped counter (uint16)
 * @param out_gi    Output buffer (must be 32 bytes)
 * @return VAX_OK on success
 */
vax_result_t vax_compute_gi(
    const uint8_t k_chain[VAX_K_CHAIN_SIZE],
    uint16_t counter,
    uint8_t out_gi[VAX_GI_SIZE]
);

/**
 * Compute SAI_n = SHA256("VAX-SAI" || prev_sai || sae || gi)
 * 
 * @param prev_sai  Previous SAI (32 bytes)
 * @param sae       Canonical JSON string (NUL-terminated)
 * @param sae_len   Length of SAE (without NUL)
 * @param gi        gi for this action (32 bytes)
 * @param out_sai   Output buffer (must be 32 bytes)
 * @return VAX_OK on success
 */
vax_result_t vax_compute_sai(
    const uint8_t prev_sai[VAX_SAI_SIZE],
    const char* sae,
    size_t sae_len,
    const uint8_t gi[VAX_GI_SIZE],
    uint8_t out_sai[VAX_SAI_SIZE]
);

/**
 * Compute genesis SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
 * 
 * @param actor_id      Actor ID string (e.g., "user123:device456")
 * @param genesis_salt  16-byte random salt
 * @param out_sai       Output buffer (must be 32 bytes)
 * @return VAX_OK on success
 */
vax_result_t vax_compute_genesis_sai(
    const char* actor_id,
    const uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE],
    uint8_t out_sai[VAX_SAI_SIZE]
);

/**
 * Canonicalize JSON to VAX-JCS format
 * 
 * @param input     Input JSON string
 * @param out_sae   Output buffer for canonical SAE
 * @param out_len   Output length (updated on success)
 * @return VAX_OK on success, VAX_ERR_INVALID_CANONICALIZATION if invalid
 */
vax_result_t vax_canonicalize_json(
    const char* input,
    char** out_sae,      // caller must free() this
    size_t* out_len
);

/**
 * Verify that a JSON string is valid VAX-JCS
 * 
 * @param sae  JSON string to verify
 * @return true if valid VAX-JCS, false otherwise
 */
bool vax_is_canonical(const char* sae);

/*============================================================================
 * High-Level API (狀態管理)
 *===========================================================================*/

/**
 * Opaque chain state handle
 */
typedef struct vax_chain vax_chain_t;

/**
 * Create a new VAX chain for an Actor
 * 
 * @param actor_id      Actor ID (e.g., "user123:device456")
 * @param k_chain       32-byte session secret
 * @param genesis_salt  16-byte genesis salt (persistent)
 * @return Chain handle, or NULL on failure
 */
vax_chain_t* vax_chain_new(
    const char* actor_id,
    const uint8_t k_chain[VAX_K_CHAIN_SIZE],
    const uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE]
);

/**
 * Free a chain handle
 */
void vax_chain_free(vax_chain_t* chain);

/**
 * Get current counter
 */
uint16_t vax_chain_get_counter(const vax_chain_t* chain);

/**
 * Get current prevSAI
 */
void vax_chain_get_prev_sai(
    const vax_chain_t* chain,
    uint8_t out_sai[VAX_SAI_SIZE]
);

/**
 * Append an action to the chain
 * 
 * This function:
 * 1. Canonicalizes the input JSON
 * 2. Derives gi
 * 3. Computes SAI
 * 4. Updates internal state (counter++, prevSAI = new SAI)
 * 
 * @param chain     Chain handle
 * @param json      Input JSON (will be canonicalized)
 * @param out_sai   Output SAI (optional, can be NULL)
 * @return VAX_OK on success
 */
vax_result_t vax_chain_append(
    vax_chain_t* chain,
    const char* json,
    uint8_t out_sai[VAX_SAI_SIZE]  // can be NULL
);

/**
 * Sync chain state (e.g., after reconnect)
 * 
 * @param chain     Chain handle
 * @param counter   Latest counter from server
 * @param prev_sai  Latest SAI from server
 */
void vax_chain_sync(
    vax_chain_t* chain,
    uint16_t counter,
    const uint8_t prev_sai[VAX_SAI_SIZE]
);

/*============================================================================
 * Verification API
 *===========================================================================*/

/**
 * Verify an action submission
 * 
 * This performs full L0 verification:
 * 1. Verify counter is expected + 1
 * 2. Verify prevSAI matches
 * 3. Verify SAE is canonical
 * 4. Recompute gi and verify
 * 5. Recompute SAI and verify
 * 
 * @param k_chain           Session secret
 * @param expected_counter  Expected counter value
 * @param expected_prev_sai Expected previous SAI
 * @param counter           Submitted counter
 * @param prev_sai          Submitted prevSAI
 * @param sae               Submitted SAE (canonical JSON)
 * @param sai               Submitted SAI
 * @return VAX_OK if verification passes
 */
vax_result_t vax_verify_action(
    const uint8_t k_chain[VAX_K_CHAIN_SIZE],
    uint16_t expected_counter,
    const uint8_t expected_prev_sai[VAX_SAI_SIZE],
    uint16_t counter,
    const uint8_t prev_sai[VAX_SAI_SIZE],
    const char* sae,
    const uint8_t sai[VAX_SAI_SIZE]
);

/*============================================================================
 * Utility Functions
 *===========================================================================*/

/**
 * Convert error code to human-readable string
 */
const char* vax_error_string(vax_result_t result);

/**
 * Convert binary to hex string (lowercase)
 * 
 * @param data      Binary data
 * @param len       Length of data
 * @param out_hex   Output buffer (must be at least len*2 + 1 bytes)
 */
void vax_bin_to_hex(
    const uint8_t* data,
    size_t len,
    char* out_hex
);

/**
 * Convert hex string to binary
 * 
 * @param hex       Hex string (must be even length)
 * @param out_data  Output buffer
 * @param out_len   Length of output (hex_len / 2)
 * @return VAX_OK on success, VAX_ERR_INVALID_INPUT if hex is invalid
 */
vax_result_t vax_hex_to_bin(
    const char* hex,
    uint8_t* out_data,
    size_t* out_len
);

#ifdef __cplusplus
}
#endif

#endif /* VAX_H */