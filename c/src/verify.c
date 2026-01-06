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
 * Verify prevSAI matches
 */
vax_result_t vax_verify_action(
    const uint8_t expected_prev_sai[VAX_SAI_SIZE],
    const uint8_t prev_sai[VAX_SAI_SIZE]
) {
    if (!expected_prev_sai || !prev_sai) {
        return VAX_ERR_INVALID_INPUT;
    }

    // Verify prevSAI matches
    if (memcmp(prev_sai, expected_prev_sai, VAX_SAI_SIZE) != 0) {
        return VAX_ERR_INVALID_PREV_SAI;
    }

    return VAX_OK;
}
