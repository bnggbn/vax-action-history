# VAX C Tests

## Running Tests

### Quick Run (All Tests)
```bash
cd c
cmake -B build -G Ninja -DBUILD_TESTS=ON
cmake --build build
ctest --test-dir build -V
```

### Run Individual Tests
```bash
cd c/build
./test_gi       # Test gi derivation
./test_sai      # Test SAI computation
./test_verify   # Test verification logic
```

### With CTest
```bash
cd c/build
ctest -V                    # Run all tests
ctest -R test_gi -V         # Run only gi tests
ctest -R test_sai -V        # Run only sai tests
ctest -R test_verify -V     # Run only verify tests
```

## Test Structure

### test_gi.c
Tests `vax_compute_gi()` - HMAC-SHA256 gi derivation

**Test Cases:**
1. ✅ gi_basic - Basic functionality
2. ✅ gi_deterministic - Same input → same output
3. ✅ gi_counter_changes - Different counter → different gi
4. ✅ gi_edge_cases - Counter boundaries (0, 65535)
5. ✅ gi_error_handling - NULL pointer checks

### test_sai.c
Tests `vax_compute_sai()` and `vax_compute_genesis_sai()` - Two-stage hash computation

**Test Cases:**
1. ✅ genesis_sai - Genesis SAI generation
2. ✅ sai_basic - Basic SAI computation
3. ✅ sai_deterministic - Deterministic SAI
4. ✅ sai_different_sae - Different SAE → different SAI
5. ✅ sai_different_prev - Different prevSAI → different SAI
6. ✅ chain_simulation - Multi-action chain
7. ✅ sai_error_handling - NULL pointer checks

### test_verify.c
Tests `vax_verify_action()` - L0 verification logic

**Test Cases:**
1. ✅ verify_valid_action - Valid action passes
2. ✅ verify_invalid_counter - Wrong counter rejected
3. ✅ verify_invalid_prev_sai - Wrong prevSAI rejected
4. ✅ verify_invalid_sai - Wrong SAI rejected
5. ✅ verify_sequence - Sequence of actions verified

## Generating Test Vectors

Run individual tests to generate reference values:

```bash
./build/test_gi > test_vectors_gi.txt
./build/test_sai > test_vectors_sai.txt
./build/test_verify > test_vectors_verify.txt
```

Then copy the hex outputs to `docs/SPECIFICATION.md` Section 13.

## Example Output

### test_gi output:
```
=== Test: vax_compute_gi (basic) ===
K_chain: 0102030405060708090a0b0c0d0e0f10...
Counter: 1
gi     : 4a5d6c... (32 bytes)
✓ gi_basic: produced 32-byte output
```

### test_sai output:
```
=== Test: vax_compute_sai (basic) ===
prevSAI: 1111111111111111... (32 bytes)
SAE: {"action":"test","value":42} (len=28)
gi     : 2222222222222222... (32 bytes)
SAI    : 8f3a2b... (32 bytes)
✓ sai_basic: produced 32-byte SAI
```

### test_verify output:
```
=== Test: vax_verify_action (valid) ===
✓ verify_valid_action: valid action accepted
```

## Adding New Tests

1. Create test function in `test_primitives.c`:
```c
void test_my_feature() {
    printf("\n=== Test: my_feature ===\n");
    // ... test code ...
    printf("✓ my_feature: passed\n");
}
```

2. Call it in `main()`:
```c
int main(void) {
    // ... existing tests ...
    test_my_feature();
    return 0;
}
```

3. Rebuild and run:
```bash
cmake --build build
./build/vax_test
```
