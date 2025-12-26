# Manual build script for VAX C library (without CMake)
# This script compiles the C library directly using clang

$OPENSSL_DIR = "C:\Program Files\OpenSSL-Win64"
$BUILD_DIR = "build"
$SRC_DIR = "src"
$INCLUDE_DIR = "include"

# Create build directory
New-Item -ItemType Directory -Force -Path $BUILD_DIR | Out-Null

Write-Host "=== Building VAX C Library ===" -ForegroundColor Green

# Compile object files
$sources = @("gi.c", "sai.c", "verify.c")
$objects = @()

foreach ($src in $sources) {
    $obj = "$BUILD_DIR/$($src -replace '\.c$', '.o')"
    $objects += $obj

    Write-Host "Compiling $src..." -ForegroundColor Cyan
    & clang -c "$SRC_DIR\$src" -o $obj `
        -I"$pwd\$INCLUDE_DIR" `
        -I"$OPENSSL_DIR\include" `
        -std=c11 -O3 -Wall -Wextra

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to compile $src" -ForegroundColor Red
        exit 1
    }
}

# Create static library
Write-Host "Creating static library libvax.a..." -ForegroundColor Cyan
& llvm-ar rcs "$BUILD_DIR\libvax.a" $objects

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== Build successful! ===" -ForegroundColor Green
    Write-Host "Output: $BUILD_DIR\libvax.a" -ForegroundColor Yellow
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}
