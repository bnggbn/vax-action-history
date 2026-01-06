#include "vax.h"
#include <openssl/rand.h>

/**
 * Compute gi = random 32 bytes (256-bit)
 *
 * Uses OpenSSL's cryptographically secure random number generator.
 */
vax_result_t vax_compute_gi(unsigned char* out_gi) {
    if (!out_gi) {
        return VAX_ERR_INVALID_INPUT;
    }

    // Generate 32 random bytes using OpenSSL CSPRNG
    if (RAND_bytes(out_gi, VAX_GI_SIZE) != 1) {
        return VAX_ERR_INVALID_INPUT;
    }

    return VAX_OK;
}
