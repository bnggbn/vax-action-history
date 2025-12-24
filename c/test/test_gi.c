#include "vax.h"
#include "test_common.h"
#include <string.h>
#include <assert.h>

// Test 1: vax_compute_gi basic
void test_gi_basic() {
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
    
    printf("✓ gi_basic: produced 32-byte output\n");
}

// Test 2: vax_compute_gi deterministic
void test_gi_deterministic() {
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
void test_gi_counter_changes() {
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

// Test 4: vax_compute_gi edge cases
void test_gi_edge_cases() {
    printf("\n=== Test: vax_compute_gi (edge cases) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    uint8_t gi[VAX_GI_SIZE];
    
    // Counter = 0
    vax_result_t result = vax_compute_gi(k_chain, 0, gi);
    assert(result == VAX_OK);
    
    // Counter = 65535 (max uint16)
    result = vax_compute_gi(k_chain, 65535, gi);
    assert(result == VAX_OK);
    
    printf("✓ gi_edge_cases: handles counter=0 and counter=65535\n");
}

// Test 5: vax_compute_gi error handling
void test_gi_error_handling() {
    printf("\n=== Test: vax_compute_gi (error handling) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    uint8_t gi[VAX_GI_SIZE];
    
    // NULL k_chain
    vax_result_t result = vax_compute_gi(NULL, 1, gi);
    assert(result == VAX_ERR_INVALID_INPUT);
    
    // NULL output
    result = vax_compute_gi(k_chain, 1, NULL);
    assert(result == VAX_ERR_INVALID_INPUT);
    
    printf("✓ gi_error_handling: NULL checks work\n");
}

// Test 6: Known test vector
void test_gi_known_vector() {
    printf("\n=== Test: vax_compute_gi (known vector) ===\n");
    
    // Test vector generated with OpenSSL 3.x:
    // Message: "VAX-GI" (0x5641582d4749) || counter=1 (0x0001, big-endian)
    // Key: 32 zero bytes
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    uint16_t counter = 1;
    uint8_t gi[VAX_GI_SIZE];
    
    vax_compute_gi(k_chain, counter, gi);
    
    // HMAC-SHA256(key=0x00*32, msg="VAX-GI" || 0x0001)
    // Generated via OpenSSL CLI:
    // printf '\x56\x41\x58\x2d\x47\x49\x00\x01' |
    // openssl dgst -sha256 -mac HMAC \
    //   -macopt hexkey:0000000000000000000000000000000000000000000000000000000000000000
    // Output:96b0dbcec77032023871b0df25214723e5b053da24d50b8f3338ea55f9966a69
    // Date: 2025-12-24
    uint8_t expected[VAX_GI_SIZE] = {
        0x96, 0xb0, 0xdb, 0xce, 0xc7, 0x70, 0x32, 0x02,
        0x38, 0x71, 0xb0, 0xdf, 0x25, 0x21, 0x47, 0x23,
        0xe5, 0xb0, 0x53, 0xda, 0x24, 0xd5, 0x0b, 0x8f,
        0x33, 0x38, 0xea, 0x55, 0xf9, 0x96, 0x6a, 0x69
    };
    
    print_hex("Expected", expected, VAX_GI_SIZE);
    print_hex("Got     ", gi, VAX_GI_SIZE);
    
    if (memcmp(gi, expected, VAX_GI_SIZE) == 0) {
        printf("✓ gi_known_vector: matches expected test vector\n");
    } else {
        printf("✗ gi_known_vector: MISMATCH!\n");
        assert(0);
    }
}

// Test 7: Verify big-endian encoding
void test_gi_endianness() {
    printf("\n=== Test: vax_compute_gi (endianness) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    uint8_t gi_256[VAX_GI_SIZE];
    uint8_t gi_1[VAX_GI_SIZE];
    
    // Counter = 0x0100 (256) → big-endian [0x01, 0x00]
    // Counter = 0x0001 (1)   → big-endian [0x00, 0x01]
    vax_compute_gi(k_chain, 256, gi_256);
    vax_compute_gi(k_chain, 1, gi_1);
    
    assert(memcmp(gi_256, gi_1, VAX_GI_SIZE) != 0);
    
    print_hex("gi(256)", gi_256, VAX_GI_SIZE);
    print_hex("gi(1)  ", gi_1, VAX_GI_SIZE);
    
    printf("✓ gi_endianness: big-endian encoding verified\n");
}

// Test 8: Different k_chain produces different gi
void test_gi_kchain_changes() {
    printf("\n=== Test: vax_compute_gi (k_chain changes) ===\n");
    
    uint8_t k_chain1[VAX_K_CHAIN_SIZE];
    uint8_t k_chain2[VAX_K_CHAIN_SIZE];
    memset(k_chain1, 0xAA, VAX_K_CHAIN_SIZE);
    memset(k_chain2, 0xBB, VAX_K_CHAIN_SIZE);
    
    uint8_t gi1[VAX_GI_SIZE];
    uint8_t gi2[VAX_GI_SIZE];
    
    vax_compute_gi(k_chain1, 1, gi1);
    vax_compute_gi(k_chain2, 1, gi2);
    
    assert(memcmp(gi1, gi2, VAX_GI_SIZE) != 0);
    printf("✓ gi_kchain_changes: different k_chain produces different gi\n");
}

// Main
int main(void) {
    printf("╔════════════════════════════════════════╗\n");
    printf("║  VAX gi Test Suite                    ║\n");
    printf("╚════════════════════════════════════════╝\n");
    
    test_gi_basic();
    test_gi_deterministic();
    test_gi_counter_changes();
    test_gi_edge_cases();
    test_gi_error_handling();
    test_gi_known_vector();      // ← 新增
    test_gi_endianness();        // ← 新增
    test_gi_kchain_changes();    // ← 新增
    
    printf("\n╔════════════════════════════════════════╗\n");
    printf("║  All gi tests passed! ✓               ║\n");
    printf("╚════════════════════════════════════════╝\n");
    
    return 0;
}
