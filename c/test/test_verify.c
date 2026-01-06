#include "vax.h"
#include "test_common.h"
#include <string.h>
#include <assert.h>

// NOTE:
// These verify tests assume SAE bytes are already canonicalized (VAX-JCS).
// Verification focuses on chain integrity, not JSON normalization.


// Test 1: verify valid action (prevSAI matches)
void test_verify_valid_action() {
    printf("\n=== Test: vax_verify_action (valid) ===\n");

    uint8_t expected_prev_sai[VAX_SAI_SIZE];
    memset(expected_prev_sai, 0xAA, VAX_SAI_SIZE);

    uint8_t prev_sai[VAX_SAI_SIZE];
    memcpy(prev_sai, expected_prev_sai, VAX_SAI_SIZE);

    // Verify
    vax_result_t result = vax_verify_action(expected_prev_sai, prev_sai);

    assert(result == VAX_OK);
    printf("✓ verify_valid_action: valid action accepted\n");
}

// Test 2: verify invalid prevSAI
void test_verify_invalid_prev_sai() {
    printf("\n=== Test: vax_verify_action (invalid prevSAI) ===\n");

    uint8_t expected_prev_sai[VAX_SAI_SIZE];
    memset(expected_prev_sai, 0xAA, VAX_SAI_SIZE);

    uint8_t wrong_prev_sai[VAX_SAI_SIZE];
    memset(wrong_prev_sai, 0xBB, VAX_SAI_SIZE);

    vax_result_t result = vax_verify_action(expected_prev_sai, wrong_prev_sai);

    assert(result == VAX_ERR_INVALID_PREV_SAI);
    printf("✓ verify_invalid_prev_sai: wrong prevSAI rejected\n");
}

// Test 3: verify NULL handling
void test_verify_null_handling() {
    printf("\n=== Test: vax_verify_action (NULL handling) ===\n");

    uint8_t prev_sai[VAX_SAI_SIZE];
    memset(prev_sai, 0xAA, VAX_SAI_SIZE);

    // NULL expected_prev_sai
    vax_result_t result = vax_verify_action(NULL, prev_sai);
    assert(result == VAX_ERR_INVALID_INPUT);

    // NULL prev_sai
    result = vax_verify_action(prev_sai, NULL);
    assert(result == VAX_ERR_INVALID_INPUT);

    printf("✓ verify_null_handling: NULL checks work\n");
}

// Main
int main(void) {
    printf("╔════════════════════════════════════════╗\n");
    printf("║  VAX Verify Test Suite                ║\n");
    printf("╚════════════════════════════════════════╝\n");

    test_verify_valid_action();
    test_verify_invalid_prev_sai();
    test_verify_null_handling();

    printf("\n╔════════════════════════════════════════╗\n");
    printf("║  All verify tests passed! ✓           ║\n");
    printf("╚════════════════════════════════════════╝\n");

    return 0;
}
