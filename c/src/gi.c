#include "vax.h"
#include <string.h>
#include <openssl/hmac.h>
#include <openssl/evp.h>

/**
 * Compute gi_n = HMAC_SHA256(k_chain, "VAX-GI" || counter)
 * 
 * Counter is encoded as big-endian uint16.
 */
vax_result_t vax_compute_gi(
    const uint8_t k_chain[VAX_K_CHAIN_SIZE],
    uint16_t counter,
    uint8_t out_gi[VAX_GI_SIZE]
) {
    if (!k_chain || !out_gi) {
        return VAX_ERR_INVALID_INPUT;
    }

    // Construct message: "VAX-GI" || counter (big-endian)
    unsigned char message[6 + 2];  // "VAX-GI" (6 bytes) + counter (2 bytes)
    memcpy(message, "VAX-GI", 6);
    
    // Encode counter as big-endian uint16
    message[6] = (counter >> 8) & 0xFF;  // High byte
    message[7] = counter & 0xFF;          // Low byte

    // Compute HMAC-SHA256
    unsigned int out_len = 0;
    unsigned char* result = HMAC(
        EVP_sha256(),
        k_chain,
        VAX_K_CHAIN_SIZE,
        message,
        sizeof(message),
        out_gi,
        &out_len
    );

    if (!result || out_len != VAX_GI_SIZE) {
        return VAX_ERR_INVALID_INPUT;
    }

    return VAX_OK;
}
