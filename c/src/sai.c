#include "vax.h"
#include <string.h>
#include <openssl/sha.h>

#define VAX_SAI_LABEL "VAX-SAI"
#define VAX_SAI_LABEL_LEN 7

vax_result_t vax_compute_sai(
    const uint8_t prev_sai[VAX_SAI_SIZE],
    const uint8_t* sae_bytes,
    size_t sae_len,
    const uint8_t gi[VAX_GI_SIZE],
    uint8_t out_sai[VAX_SAI_SIZE]
) {
    if (!prev_sai || !sae_bytes || !gi || !out_sai) {
        return VAX_ERR_INVALID_INPUT;
    }

    // sae_hash = SHA256(SAE)
    uint8_t sae_hash[VAX_SAI_SIZE];
    SHA256(sae_bytes, sae_len, sae_hash);

    // message = "VAX-SAI" || prevSAI || sae_hash || gi
    uint8_t message[VAX_SAI_LABEL_LEN + VAX_SAI_SIZE + VAX_SAI_SIZE + VAX_GI_SIZE];
    size_t off = 0;

    memcpy(message + off, VAX_SAI_LABEL, VAX_SAI_LABEL_LEN); off += VAX_SAI_LABEL_LEN;
    memcpy(message + off, prev_sai, VAX_SAI_SIZE);           off += VAX_SAI_SIZE;
    memcpy(message + off, sae_hash, VAX_SAI_SIZE);           off += VAX_SAI_SIZE;
    memcpy(message + off, gi, VAX_GI_SIZE);                  off += VAX_GI_SIZE;

    SHA256(message, sizeof(message), out_sai);
    return VAX_OK;
}


/**
 * Compute genesis prevSAI from genesis_salt
 * 
 * Genesis prevSAI = SHA256("VAX-GENESIS" || genesis_salt)
 */
vax_result_t vax_compute_genesis_sai(
    const char* actor_id,
    const uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE],
    uint8_t out_genesis_sai[VAX_SAI_SIZE]
) {
    if (!actor_id || !genesis_salt || !out_genesis_sai) {
        return VAX_ERR_INVALID_INPUT;
    }

    size_t actor_id_len = strlen(actor_id);
    
    // Message: "VAX-GENESIS" || actor_id || genesis_salt
    size_t message_len = 11 + actor_id_len + VAX_GENESIS_SALT_SIZE;
    unsigned char* message = malloc(message_len);
    if (!message) {
        return VAX_ERR_OUT_OF_MEMORY;
    }

    memcpy(message, "VAX-GENESIS", 11);
    memcpy(message + 11, actor_id, actor_id_len);
    memcpy(message + 11 + actor_id_len, genesis_salt, VAX_GENESIS_SALT_SIZE);

    SHA256(message, message_len, out_genesis_sai);
    free(message);
    
    return VAX_OK;
}