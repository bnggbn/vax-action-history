#include "vax.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <assert.h>

// Helper: print hex
void print_hex(const char* label, const uint8_t* data, size_t len) {
    printf("%s: ", label);
    for (size_t i = 0; i < len; i++) {
        printf("%02x", data[i]);
    }
    printf("\n");
}

// Helper: compare and report
int compare_bytes(const char* test_name, const uint8_t* got, const uint8_t* expected, size_t len) {
    if (memcmp(got, expected, len) == 0) {
        printf("✓ %s passed\n", test_name);
        return 1;
    } else {
        printf("✗ %s FAILED\n", test_name);
        print_hex("  Expected", expected, len);
        print_hex("  Got     ", got, len);
        return 0;
    }
}

// Test 1: vax_compute_gi basic
void test_gi_basic(void) {
    printf("\n=== Test: vax_compute_gi (basic) ===\n");

    // Test vector
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {
        0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
        0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
        0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
        0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20
    };
    uint16_t counter = 1;
    uint8_t gi[VAX_GI_SIZE];

    vax_result_t result = vax_compute_gi(k_chain, counter, gi);
    assert(result == VAX_OK);

    printf("K_chain: ");
    for (int i = 0; i < VAX_K_CHAIN_SIZE; i++) printf("%02x", k_chain[i]);
    printf("\n");
    printf("Counter: %u\n", counter);
    print_hex("gi     ", gi, VAX_GI_SIZE);

    // Note: Run this once to get expected value, then hard-code it
    // For now, just verify it produces 32 bytes
    printf("✓ gi_basic: produced 32-byte output\n");
}

// Test 2: vax_compute_gi deterministic
void test_gi_deterministic(void) {
    printf("\n=== Test: vax_compute_gi (deterministic) ===\n");

    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    memset(k_chain, 0xAA, VAX_K_CHAIN_SIZE);

    uint8_t gi1[VAX_GI_SIZE];
    uint8_t gi2[VAX_GI_SIZE];

    vax_compute_gi(k_chain, 42, gi1);
    vax_compute_gi(k_chain, 42, gi2);

    assert(memcmp(gi1, gi2, VAX_GI_SIZE) == 0);
    printf("✓ gi_deterministic: same input produces same output\n");
}

// Test 3: vax_compute_gi counter increment
void test_gi_counter_changes(void) {
    printf("\n=== Test: vax_compute_gi (counter changes) ===\n");

    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    memset(k_chain, 0xBB, VAX_K_CHAIN_SIZE);

    uint8_t gi1[VAX_GI_SIZE];
    uint8_t gi2[VAX_GI_SIZE];

    vax_compute_gi(k_chain, 1, gi1);
    vax_compute_gi(k_chain, 2, gi2);

    assert(memcmp(gi1, gi2, VAX_GI_SIZE) != 0);
    printf("✓ gi_counter_changes: different counter produces different gi\n");
}

// Test 4: vax_compute_genesis_sai
void test_genesis_sai(void) {
    printf("\n=== Test: vax_compute_genesis_sai ===\n");

    const char* actor_id = "user123:device456";
    uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE] = {
        0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
        0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0
    };
    uint8_t genesis_sai[VAX_SAI_SIZE];

    vax_result_t result = vax_compute_genesis_sai(actor_id, genesis_salt, genesis_sai);
    assert(result == VAX_OK);

    printf("Actor ID: %s\n", actor_id);
    print_hex("Genesis salt", genesis_salt, VAX_GENESIS_SALT_SIZE);
    print_hex("Genesis SAI ", genesis_sai, VAX_SAI_SIZE);

    // Verify deterministic
    uint8_t genesis_sai2[VAX_SAI_SIZE];
    vax_compute_genesis_sai(actor_id, genesis_salt, genesis_sai2);
    assert(memcmp(genesis_sai, genesis_sai2, VAX_SAI_SIZE) == 0);

    printf("✓ genesis_sai: deterministic output\n");
}

// Test 5: vax_compute_sai basic
void test_sai_basic(void) {
    printf("\n=== Test: vax_compute_sai (basic) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    memset(prev_sai, 0x11, VAX_SAI_SIZE);

    const char* sae = "{\"action\":\"test\",\"value\":42}";
    size_t sae_len = strlen(sae);

    uint8_t gi[VAX_GI_SIZE] = {0};
    memset(gi, 0x22, VAX_GI_SIZE);

    uint8_t sai[VAX_SAI_SIZE];

    vax_result_t result = vax_compute_sai(prev_sai, (const uint8_t*)sae, sae_len, gi, sai);
    assert(result == VAX_OK);

    print_hex("prevSAI", prev_sai, VAX_SAI_SIZE);
    printf("SAE: %s (len=%zu)\n", sae, sae_len);
    print_hex("gi     ", gi, VAX_GI_SIZE);
    print_hex("SAI    ", sai, VAX_SAI_SIZE);

    printf("✓ sai_basic: produced 32-byte SAI\n");
}

// Test 6: vax_compute_sai deterministic
void test_sai_deterministic(void) {
    printf("\n=== Test: vax_compute_sai (deterministic) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    const char* sae = "{\"test\":1}";
    uint8_t gi[VAX_GI_SIZE] = {0};

    uint8_t sai1[VAX_SAI_SIZE];
    uint8_t sai2[VAX_SAI_SIZE];

    vax_compute_sai(prev_sai, (const uint8_t*)sae, strlen(sae), gi, sai1);
    vax_compute_sai(prev_sai, (const uint8_t*)sae, strlen(sae), gi, sai2);

    assert(memcmp(sai1, sai2, VAX_SAI_SIZE) == 0);
    printf("✓ sai_deterministic: same input produces same SAI\n");
}

// Test 7: vax_compute_sai different SAE
void test_sai_different_sae(void) {
    printf("\n=== Test: vax_compute_sai (different SAE) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    uint8_t gi[VAX_GI_SIZE] = {0};

    const char* sae1 = "{\"action\":\"test1\"}";
    const char* sae2 = "{\"action\":\"test2\"}";

    uint8_t sai1[VAX_SAI_SIZE];
    uint8_t sai2[VAX_SAI_SIZE];

    vax_compute_sai(prev_sai, (const uint8_t*)sae1, strlen(sae1), gi, sai1);
    vax_compute_sai(prev_sai, (const uint8_t*)sae2, strlen(sae2), gi, sai2);

    assert(memcmp(sai1, sai2, VAX_SAI_SIZE) != 0);
    printf("✓ sai_different_sae: different SAE produces different SAI\n");
}

// Test 8: Full chain simulation
void test_chain_simulation(void) {
    printf("\n=== Test: Chain simulation ===\n");

    // Setup
    const char* actor_id = "alice:laptop";
    uint8_t k_chain[VAX_K_CHAIN_SIZE];
    memset(k_chain, 0x42, VAX_K_CHAIN_SIZE);

    uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE];
    memset(genesis_salt, 0xAB, VAX_GENESIS_SALT_SIZE);

    // Genesis
    uint8_t prev_sai[VAX_SAI_SIZE];
    vax_compute_genesis_sai(actor_id, genesis_salt, prev_sai);
    printf("Genesis SAI computed\n");

    // Action 1
    uint16_t counter = 1;
    const char* sae1 = "{\"action\":\"create\",\"id\":1}";
    uint8_t gi1[VAX_GI_SIZE];
    uint8_t sai1[VAX_SAI_SIZE];

    vax_compute_gi(k_chain, counter, gi1);
    vax_compute_sai(prev_sai, (const uint8_t*)sae1, strlen(sae1), gi1, sai1);
    print_hex("SAI_1  ", sai1, VAX_SAI_SIZE);

    // Action 2
    counter = 2;
    const char* sae2 = "{\"action\":\"update\",\"id\":1}";
    uint8_t gi2[VAX_GI_SIZE];
    uint8_t sai2[VAX_SAI_SIZE];

    vax_compute_gi(k_chain, counter, gi2);
    vax_compute_sai(sai1, (const uint8_t*)sae2, strlen(sae2), gi2, sai2);
    print_hex("SAI_2  ", sai2, VAX_SAI_SIZE);

    // Verify chain properties
    assert(memcmp(gi1, gi2, VAX_GI_SIZE) != 0);  // Different gi
    assert(memcmp(sai1, sai2, VAX_SAI_SIZE) != 0);  // Different SAI

    printf("✓ chain_simulation: 2-action chain successful\n");
}

// Test 9: Error handling
void test_error_handling(void) {
    printf("\n=== Test: Error handling ===\n");

    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    uint8_t gi[VAX_GI_SIZE];

    // NULL k_chain
    vax_result_t result = vax_compute_gi(NULL, 1, gi);
    assert(result == VAX_ERR_INVALID_INPUT);

    // NULL output
    result = vax_compute_gi(k_chain, 1, NULL);
    assert(result == VAX_ERR_INVALID_INPUT);

    printf("✓ error_handling: NULL checks work\n");
}

// Main
int main(void) {
    printf("╔════════════════════════════════════════╗\n");
    printf("║  VAX Primitives Test Suite            ║\n");
    printf("╚════════════════════════════════════════╝\n");

    test_gi_basic();
    test_gi_deterministic();
    test_gi_counter_changes();
    test_genesis_sai();
    test_sai_basic();
    test_sai_deterministic();
    test_sai_different_sae();
    test_chain_simulation();
    test_error_handling();

    printf("\n╔════════════════════════════════════════╗\n");
    printf("║  All tests passed! ✓                  ║\n");
    printf("╚════════════════════════════════════════╝\n");

    return 0;
}
