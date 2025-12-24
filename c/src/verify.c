#include "vax.h"
#include <string.h>

/**
 * Verify an action submission (crypto only)
 * 
 * NOTE: This does NOT verify SAE canonicalization.
 * Canonicalization check should be done at the application layer
 * (Go server or JS/TS frontend).
 * 
 * This performs:
 * 1. Verify counter is expected + 1
 * 2. Verify prevSAI matches
 * 3. Recompute gi
 * 4. Recompute SAI and verify
 */
vax_result_t vax_verify_action(
    const uint8_t k_chain[VAX_K_CHAIN_SIZE],
    uint16_t expected_counter,
    const uint8_t expected_prev_sai[VAX_SAI_SIZE],
    uint16_t counter,
    const uint8_t prev_sai[VAX_SAI_SIZE],
    const uint8_t* sae_bytes,
    size_t sae_len,
    const uint8_t sai[VAX_SAI_SIZE]
) {
    // Step 0: Counter overflow check
    if (expected_counter == UINT16_MAX) {
        return VAX_ERR_COUNTER_OVERFLOW;
    }

    // Step 1: Verify counter is expected + 1
    if (counter != expected_counter + 1) {
        return VAX_ERR_INVALID_COUNTER;
    }

    // Step 2: Verify prevSAI matches
    if (memcmp(prev_sai, expected_prev_sai, VAX_SAI_SIZE) != 0) {
        return VAX_ERR_INVALID_PREV_SAI;
    }

    // Step 3: Recompute gi
    uint8_t computed_gi[VAX_GI_SIZE];
    vax_result_t result = vax_compute_gi(k_chain, counter, computed_gi);
    if (result != VAX_OK) {
        return result;
    }

    // Step 4: Recompute SAI and verify
    uint8_t computed_sai[VAX_SAI_SIZE];
    result = vax_compute_sai(prev_sai, sae_bytes, sae_len, computed_gi, computed_sai);
    if (result != VAX_OK) {
        return result;
    }

    // Verify SAI matches
    if (memcmp(sai, computed_sai, VAX_SAI_SIZE) != 0) {
        return VAX_ERR_SAI_MISMATCH;
    }

    return VAX_OK;
}
