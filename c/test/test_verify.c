#include "vax.h"
#include "test_common.h"
#include <string.h>
#include <assert.h>

// Test 1: verify valid action
void test_verify_valid_action() {
    printf("\n=== Test: vax_verify_action (valid) ===\n");
    
    // Setup
    uint8_t k_chain[VAX_K_CHAIN_SIZE];
    memset(k_chain, 0x42, VAX_K_CHAIN_SIZE);
    
    uint16_t expected_counter = 0;
    uint8_t expected_prev_sai[VAX_SAI_SIZE];
    memset(expected_prev_sai, 0xAA, VAX_SAI_SIZE);
    
    // Create action
    uint16_t counter = 1;
    uint8_t prev_sai[VAX_SAI_SIZE];
    memcpy(prev_sai, expected_prev_sai, VAX_SAI_SIZE);
    
    const char* sae = "{\"action\":\"test\"}";
    
    // Compute gi and sai
    uint8_t gi[VAX_GI_SIZE];
    uint8_t sai[VAX_SAI_SIZE];
    vax_compute_gi(k_chain, counter, gi);
    vax_compute_sai(prev_sai, (const uint8_t*)sae, strlen(sae), gi, sai);
    
    // Verify
    vax_result_t result = vax_verify_action(
        k_chain,
        expected_counter,
        expected_prev_sai,
        counter,
        prev_sai,
        sae,
        sai
    );
    
    assert(result == VAX_OK);
    printf("✓ verify_valid_action: valid action accepted\n");
}

// Test 2: verify invalid counter
void test_verify_invalid_counter() {
    printf("\n=== Test: vax_verify_action (invalid counter) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    uint8_t expected_prev_sai[VAX_SAI_SIZE] = {0};
    uint8_t prev_sai[VAX_SAI_SIZE] = {0};
    const char* sae = "{\"test\":1}";
    uint8_t sai[VAX_SAI_SIZE] = {0};
    
    // Counter is not expected + 1
    vax_result_t result = vax_verify_action(
        k_chain,
        5,              // expected = 5
        expected_prev_sai,
        10,             // submitted = 10 (should be 6)
        prev_sai,
        sae,
        sai
    );
    
    assert(result == VAX_ERR_INVALID_COUNTER);
    printf("✓ verify_invalid_counter: wrong counter rejected\n");
}

// Test 3: verify invalid prevSAI
void test_verify_invalid_prev_sai() {
    printf("\n=== Test: vax_verify_action (invalid prevSAI) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE] = {0};
    
    uint8_t expected_prev_sai[VAX_SAI_SIZE];
    memset(expected_prev_sai, 0xAA, VAX_SAI_SIZE);
    
    uint8_t wrong_prev_sai[VAX_SAI_SIZE];
    memset(wrong_prev_sai, 0xBB, VAX_SAI_SIZE);
    
    const char* sae = "{\"test\":1}";
    uint8_t sai[VAX_SAI_SIZE] = {0};
    
    vax_result_t result = vax_verify_action(
        k_chain,
        0,
        expected_prev_sai,
        1,
        wrong_prev_sai,  // Different prevSAI
        sae,
        sai
    );
    
    assert(result == VAX_ERR_INVALID_PREV_SAI);
    printf("✓ verify_invalid_prev_sai: wrong prevSAI rejected\n");
}

// Test 4: verify invalid SAI
void test_verify_invalid_sai() {
    printf("\n=== Test: vax_verify_action (invalid SAI) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE];
    memset(k_chain, 0x42, VAX_K_CHAIN_SIZE);
    
    uint8_t expected_prev_sai[VAX_SAI_SIZE];
    memset(expected_prev_sai, 0xAA, VAX_SAI_SIZE);
    
    uint16_t counter = 1;
    uint8_t prev_sai[VAX_SAI_SIZE];
    memcpy(prev_sai, expected_prev_sai, VAX_SAI_SIZE);
    
    const char* sae = "{\"action\":\"test\"}";
    
    // Compute correct gi but provide wrong SAI
    uint8_t gi[VAX_GI_SIZE];
    vax_compute_gi(k_chain, counter, gi);
    
    uint8_t wrong_sai[VAX_SAI_SIZE];
    memset(wrong_sai, 0xFF, VAX_SAI_SIZE);  // Wrong SAI
    
    vax_result_t result = vax_verify_action(
        k_chain,
        0,
        expected_prev_sai,
        counter,
        prev_sai,
        sae,
        wrong_sai
    );
    
    assert(result == VAX_ERR_SAI_MISMATCH);
    printf("✓ verify_invalid_sai: wrong SAI rejected\n");
}

// Test 5: verify sequence of actions
void test_verify_sequence() {
    printf("\n=== Test: vax_verify_action (sequence) ===\n");
    
    uint8_t k_chain[VAX_K_CHAIN_SIZE];
    memset(k_chain, 0x42, VAX_K_CHAIN_SIZE);
    
    // Start with genesis
    const char* actor_id = "test:device";
    uint8_t genesis_salt[VAX_GENESIS_SALT_SIZE];
    memset(genesis_salt, 0xCD, VAX_GENESIS_SALT_SIZE);
    
    uint8_t genesis_sai[VAX_SAI_SIZE];
    vax_compute_genesis_sai(actor_id, genesis_salt, genesis_sai);
    
    // Action 1
    uint16_t counter = 1;
    const char* sae1 = "{\"action\":\"create\"}";
    uint8_t gi1[VAX_GI_SIZE];
    uint8_t sai1[VAX_SAI_SIZE];
    
    vax_compute_gi(k_chain, counter, gi1);
    vax_compute_sai(genesis_sai, (const uint8_t*)sae1, strlen(sae1), gi1, sai1);
    
    vax_result_t result = vax_verify_action(
        k_chain,
        0,           // expected counter
        genesis_sai, // expected prevSAI
        counter,
        genesis_sai,
        sae1,
        sai1
    );
    assert(result == VAX_OK);
    printf("Action 1 verified ✓\n");
    
    // Action 2
    counter = 2;
    const char* sae2 = "{\"action\":\"update\"}";
    uint8_t gi2[VAX_GI_SIZE];
    uint8_t sai2[VAX_SAI_SIZE];
    
    vax_compute_gi(k_chain, counter, gi2);
    vax_compute_sai(sai1, (const uint8_t*)sae2, strlen(sae2), gi2, sai2);
    
    result = vax_verify_action(
        k_chain,
        1,     // expected counter
        sai1,  // expected prevSAI
        counter,
        sai1,
        sae2,
        sai2
    );
    assert(result == VAX_OK);
    printf("Action 2 verified ✓\n");
    
    printf("✓ verify_sequence: sequence of 2 actions verified\n");
}

// Main
int main(void) {
    printf("╔════════════════════════════════════════╗\n");
    printf("║  VAX Verify Test Suite                ║\n");
    printf("╚════════════════════════════════════════╝\n");
    
    test_verify_valid_action();
    test_verify_invalid_counter();
    test_verify_invalid_prev_sai();
    test_verify_invalid_sai();
    test_verify_sequence();
    
    printf("\n╔════════════════════════════════════════╗\n");
    printf("║  All verify tests passed! ✓           ║\n");
    printf("╚════════════════════════════════════════╝\n");
    
    return 0;
}
