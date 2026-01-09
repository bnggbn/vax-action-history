#include "vax.h"
#include "test_common.h"
#include <string.h>
#include <assert.h>
// NOTE:
// These tests assume SAE bytes are already canonicalized (VAX-JCS).
// This suite verifies hashing / chaining semantics, not JSON canonicalization.

// Test 1: vax_compute_genesis_sai
void test_genesis_sai(void) {
    printf("\n=== Test: vax_compute_genesis_sai ===\n");

    const char* actor_id = "user123:device456";
    uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE] = {
        0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
        0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0
    };

    uint8_t genesis_sai[VAX_SAI_SIZE];
    vax_result_t result =
        vax_compute_genesis_sai(actor_id, genesis_salt, genesis_sai);
    assert(result == VAX_OK);

    print_hex("Genesis SAI", genesis_sai, VAX_SAI_SIZE);

    // ---- Golden vector verification ----
    // Generated via OpenSSL CLI:
    //
    // printf '\x56\x41\x58\x2d\x47\x45\x4e\x45\x53\x49\x53'\
    //        '\x75\x73\x65\x72\x31\x32\x33\x3a\x64\x65\x76\x69\x63\x65\x34\x35\x36'\
    //        '\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0' |
    // openssl dgst -sha256
    // Output:afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3
    // Date : 2025-12-24
    uint8_t expected[VAX_SAI_SIZE] = {
        0xaf, 0xc5, 0x07, 0x28, 0xcd, 0x79, 0xe8, 0x05,
        0xa8, 0xae, 0x06, 0x87, 0x5a, 0x1d, 0xdf, 0x78,
        0xca, 0x11, 0xb0, 0xd5, 0x6e, 0xc3, 0x00, 0xb1,
        0x60, 0xfb, 0x71, 0xf5, 0x0c, 0xe6, 0x58, 0xc3
    };


    print_hex("Expected   ", expected, VAX_SAI_SIZE);
    assert(memcmp(genesis_sai, expected, VAX_SAI_SIZE) == 0);

    printf("✓ genesis_sai: matches golden vector\n");
}

// Test 2: vax_compute_sai basic
void test_sai_basic(void) {
    printf("\n=== Test: vax_compute_sai (basic) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    memset(prev_sai, 0x11, VAX_SAI_SIZE);

    const char* sae = "{\"action\":\"test\",\"value\":42}";
    size_t sae_len = strlen(sae);

    uint8_t sai[VAX_SAI_SIZE];

    vax_result_t result = vax_compute_sai(prev_sai, (const uint8_t*)sae, sae_len, sai);
    assert(result == VAX_OK);

    print_hex("prevSAI", prev_sai, VAX_SAI_SIZE);
    printf("SAE: %s (len=%zu)\n", sae, sae_len);
    print_hex("SAI    ", sai, VAX_SAI_SIZE);

    printf("✓ sai_basic: produced 32-byte SAI\n");
}

// Test 3: vax_compute_sai randomness (gi is random internally)
void test_sai_randomness(void) {
    printf("\n=== Test: vax_compute_sai (randomness) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    const char* sae = "{\"test\":1}";

    uint8_t sai1[VAX_SAI_SIZE];
    uint8_t sai2[VAX_SAI_SIZE];

    // Since gi is generated randomly inside, same inputs produce different outputs
    vax_compute_sai(prev_sai, (const uint8_t*)sae, strlen(sae), sai1);
    vax_compute_sai(prev_sai, (const uint8_t*)sae, strlen(sae), sai2);

    assert(memcmp(sai1, sai2, VAX_SAI_SIZE) != 0);
    printf("✓ sai_randomness: same input produces different SAI (random gi)\n");
}

// Test 4: Full chain simulation
void test_chain_simulation(void) {
    printf("\n=== Test: Chain simulation ===\n");

    // Setup
    const char* actor_id = "alice:laptop";

    uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE];
    memset(genesis_salt, 0xAB, VAX_GENESIS_SALT_SIZE);

    // Genesis
    uint8_t prev_sai[VAX_SAI_SIZE];
    vax_compute_genesis_sai(actor_id, genesis_salt, prev_sai);
    printf("Genesis SAI computed\n");

    // Action 1
    const char* sae1 = "{\"action\":\"create\",\"id\":1}";
    uint8_t sai1[VAX_SAI_SIZE];

    vax_compute_sai(prev_sai, (const uint8_t*)sae1, strlen(sae1), sai1);
    print_hex("SAI_1  ", sai1, VAX_SAI_SIZE);

    // Action 2
    const char* sae2 = "{\"action\":\"update\",\"id\":1}";
    uint8_t sai2[VAX_SAI_SIZE];

    vax_compute_sai(sai1, (const uint8_t*)sae2, strlen(sae2), sai2);
    print_hex("SAI_2  ", sai2, VAX_SAI_SIZE);

    // Verify chain properties
    assert(memcmp(sai1, sai2, VAX_SAI_SIZE) != 0);  // Different SAI

    printf("✓ chain_simulation: 2-action chain successful\n");
}

// Test 5: SAI error handling
void test_sai_error_handling(void) {
    printf("\n=== Test: vax_compute_sai (error handling) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    const char* sae = "{\"test\":1}";
    uint8_t sai[VAX_SAI_SIZE];

    // NULL prev_sai
    vax_result_t result = vax_compute_sai(NULL, (const uint8_t*)sae, strlen(sae), sai);
    assert(result == VAX_ERR_INVALID_INPUT);

    // NULL sae
    result = vax_compute_sai(prev_sai, NULL, 10, sai);
    assert(result == VAX_ERR_INVALID_INPUT);

    // NULL output
    result = vax_compute_sai(prev_sai, (const uint8_t*)sae, strlen(sae), NULL);
    assert(result == VAX_ERR_INVALID_INPUT);

    printf("✓ sai_error_handling: NULL checks work\n");
}




// Main
int main(void) {
    printf("╔════════════════════════════════════════╗\n");
    printf("║  VAX SAI Test Suite                   ║\n");
    printf("╚════════════════════════════════════════╝\n");

    test_genesis_sai();
    test_sai_basic();
    test_sai_randomness();
    test_chain_simulation();
    test_sai_error_handling();

    printf("\n╔════════════════════════════════════════╗\n");
    printf("║  All SAI tests passed! ✓              ║\n");
    printf("╚════════════════════════════════════════╝\n");

    return 0;
}
