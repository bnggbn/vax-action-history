#ifndef TEST_COMMON_H
#define TEST_COMMON_H

#include <stdio.h>
#include <stdint.h>
#include <stddef.h>

// Helper: print hex
static inline void print_hex(const char* label, const uint8_t* data, size_t len) {
    printf("%s: ", label);
    for (size_t i = 0; i < len; i++) {
        printf("%02x", data[i]);
    }
    printf("\n");
}

// Helper: compare and report
static inline int compare_bytes(const char* test_name, const uint8_t* got, const uint8_t* expected, size_t len) {
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

#endif // TEST_COMMON_H
