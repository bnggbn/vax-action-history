#include "vax.h"
#include "test_common.h"
#include <string.h>
#include <assert.h>

// Forward declaration for internal function (defined in gi.c)
vax_result_t vax_compute_gi(uint8_t* out_gi);

// Test 1: vax_compute_gi basic - produces 32 bytes
void test_gi_basic() {
    printf("\n=== Test: vax_compute_gi (basic) ===\n");

    uint8_t gi[VAX_GI_SIZE];

    vax_result_t result = vax_compute_gi(gi);
    assert(result == VAX_OK);

    print_hex("gi", gi, VAX_GI_SIZE);

    printf("✓ gi_basic: produced 32-byte output\n");
}

// Test 2: vax_compute_gi randomness - each call produces different output
void test_gi_randomness() {
    printf("\n=== Test: vax_compute_gi (randomness) ===\n");

    uint8_t gi1[VAX_GI_SIZE];
    uint8_t gi2[VAX_GI_SIZE];

    vax_compute_gi(gi1);
    vax_compute_gi(gi2);

    // Two random values should be different (extremely high probability)
    assert(memcmp(gi1, gi2, VAX_GI_SIZE) != 0);

    print_hex("gi1", gi1, VAX_GI_SIZE);
    print_hex("gi2", gi2, VAX_GI_SIZE);

    printf("✓ gi_randomness: each call produces different output\n");
}

// Test 3: vax_compute_gi error handling
void test_gi_error_handling() {
    printf("\n=== Test: vax_compute_gi (error handling) ===\n");

    // NULL output
    vax_result_t result = vax_compute_gi(NULL);
    assert(result == VAX_ERR_INVALID_INPUT);

    printf("✓ gi_error_handling: NULL check works\n");
}

// Test 4: vax_compute_gi non-zero output
void test_gi_non_zero() {
    printf("\n=== Test: vax_compute_gi (non-zero) ===\n");

    uint8_t gi[VAX_GI_SIZE];
    uint8_t zeros[VAX_GI_SIZE] = {0};

    vax_compute_gi(gi);

    // Random output should not be all zeros (extremely high probability)
    assert(memcmp(gi, zeros, VAX_GI_SIZE) != 0);

    printf("✓ gi_non_zero: output is not all zeros\n");
}

// Test 5: vax_compute_gi multiple calls
void test_gi_multiple_calls() {
    printf("\n=== Test: vax_compute_gi (multiple calls) ===\n");

    const int NUM_CALLS = 10;
    uint8_t gis[10][VAX_GI_SIZE];

    // Generate multiple gi values
    for (int i = 0; i < NUM_CALLS; i++) {
        vax_result_t result = vax_compute_gi(gis[i]);
        assert(result == VAX_OK);
    }

    // Verify all are unique
    for (int i = 0; i < NUM_CALLS; i++) {
        for (int j = i + 1; j < NUM_CALLS; j++) {
            assert(memcmp(gis[i], gis[j], VAX_GI_SIZE) != 0);
        }
    }

    printf("✓ gi_multiple_calls: all %d calls produced unique values\n", NUM_CALLS);
}

// Main
int main(void) {
    printf("╔════════════════════════════════════════╗\n");
    printf("║  VAX gi Test Suite                    ║\n");
    printf("╚════════════════════════════════════════╝\n");

    test_gi_basic();
    test_gi_randomness();
    test_gi_error_handling();
    test_gi_non_zero();
    test_gi_multiple_calls();

    printf("\n╔════════════════════════════════════════╗\n");
    printf("║  All gi tests passed! ✓               ║\n");
    printf("╚════════════════════════════════════════╝\n");

    return 0;
}
