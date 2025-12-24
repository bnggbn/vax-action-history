# VAX C Library - Build Instructions

## Prerequisites

### Windows

```powershell
# Install LLVM (Clang)
winget install LLVM.LLVM

# Install CMake
winget install Kitware.CMake

# Install Ninja (optional but recommended)
winget install Ninja-build.Ninja

# Install OpenSSL (choose one):
# Option 1: vcpkg
vcpkg install openssl:x64-windows

# Option 2: MSYS2
pacman -S mingw-w64-x86_64-openssl
```

### Linux/macOS

```bash
# Ubuntu/Debian
sudo apt-get install clang cmake ninja-build libssl-dev

# macOS
brew install llvm cmake ninja openssl
```

## Build

### Quick Build

```bash
cd c

# Configure
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release

# Build
cmake --build build

# Output: build/libvax.a
```

### Debug Build (with Sanitizers)

```bash
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug
cmake --build build

# This enables:
# - AddressSanitizer (memory errors)
# - UndefinedBehaviorSanitizer (UB detection)
```

### With Tests

```bash
cmake -B build -G Ninja -DBUILD_TESTS=ON
cmake --build build
ctest --test-dir build
```

### VS Code Integration

1. Open workspace in VS Code
2. Install recommended extensions (prompt will appear)
3. Press `Ctrl+Shift+P` → "CMake: Configure"
4. Press `F7` to build

## Verify Installation

```bash
# Check tools
clang --version      # Should show LLVM version
cmake --version      # Should show CMake 3.15+
ninja --version      # Should show Ninja version

# Test compilation
cd c/build
ls libvax.a          # Should exist after build
```

## Clean Build

```bash
rm -rf c/build
cmake -B c/build -G Ninja
cmake --build c/build
```

## Troubleshooting

### "Cannot find OpenSSL"

```bash
# Specify OpenSSL path manually
cmake -B build -DOPENSSL_ROOT_DIR=/path/to/openssl
```

### "clang: command not found"

```bash
# Check LLVM is in PATH
where clang         # Windows
which clang         # Linux/macOS

# Or specify manually
cmake -B build -DCMAKE_C_COMPILER=/path/to/clang
```

### Windows: "undefined reference to OpenSSL functions"

Use vcpkg toolchain:
```powershell
cmake -B build -G Ninja `
  -DCMAKE_TOOLCHAIN_FILE="C:/path/to/vcpkg/scripts/buildsystems/vcpkg.cmake"
```
