#include "vax.h"
#include <string.h>

/**
 * Verify an action submission
 * 
 * This performs full L0 verification:
 * 1. Verify counter is expected + 1
 * 2. Verify prevSAI matches
 * 3. Verify SAE is canonical
 * 4. Recompute gi and verify
 * 5. Recompute SAI and verify
 */
vax_result_t vax_verify_action(
    const uint8_t k_chain[VAX_K_CHAIN_SIZE],
    uint16_t expected_counter,
    const uint8_t expected_prev_sai[VAX_SAI_SIZE],
    uint16_t counter,
    const uint8_t prev_sai[VAX_SAI_SIZE],
    const char* sae,
    const uint8_t sai[VAX_SAI_SIZE]
) {
    // Step 1: Verify counter is expected + 1
    if (counter != expected_counter + 1) {
        return VAX_ERR_INVALID_COUNTER;
    }

    // Step 2: Verify prevSAI matches
    if (memcmp(prev_sai, expected_prev_sai, VAX_SAI_SIZE) != 0) {
        return VAX_ERR_INVALID_PREV_SAI;
    }

    // Step 3: Verify SAE is canonical
    if (!vax_is_canonical(sae)) {
        return VAX_ERR_INVALID_CANONICALIZATION;
    }

    // Step 4: Recompute gi and verify
    uint8_t computed_gi[VAX_GI_SIZE];
    vax_result_t result = vax_compute_gi(k_chain, counter, computed_gi);
    if (result != VAX_OK) {
        return result;
    }

    // Note: We don't have the submitted gi to compare directly.
    // Instead, we'll verify by recomputing SAI with our computed gi.

    // Step 5: Recompute SAI and verify
    uint8_t computed_sai[VAX_SAI_SIZE];
    size_t sae_len = strlen(sae);
    result = vax_compute_sai(prev_sai, (const uint8_t*)sae, sae_len, computed_gi, computed_sai);
    if (result != VAX_OK) {
        return result;
    }

    // Verify SAI matches
    if (memcmp(sai, computed_sai, VAX_SAI_SIZE) != 0) {
        return VAX_ERR_SAI_MISMATCH;
    }

    return VAX_OK;
}
