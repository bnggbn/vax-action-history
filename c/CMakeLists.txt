cmake_minimum_required(VERSION 3.15)

# Use Clang if available
if(NOT CMAKE_C_COMPILER)
    find_program(CLANG_EXECUTABLE NAMES clang)
    if(CLANG_EXECUTABLE)
        set(CMAKE_C_COMPILER ${CLANG_EXECUTABLE})
        message(STATUS "Using Clang: ${CLANG_EXECUTABLE}")
    endif()
endif()

project(vax C)

# C11 standard
set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS OFF)

# Compiler flags
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -Wextra -Wpedantic")

# Clang-specific flags
if(CMAKE_C_COMPILER_ID MATCHES "Clang")
    set(CMAKE_C_FLAGS_DEBUG "-g -O0 -fsanitize=address -fsanitize=undefined")
    set(CMAKE_C_FLAGS_RELEASE "-O3 -flto")
    message(STATUS "Clang detected - enabling sanitizers in Debug mode")
endif()

# GCC-specific flags (fallback)
if(CMAKE_C_COMPILER_ID MATCHES "GNU")
    set(CMAKE_C_FLAGS_DEBUG "-g -O0")
    set(CMAKE_C_FLAGS_RELEASE "-O3")
endif()

# Find OpenSSL
find_package(OpenSSL REQUIRED)
message(STATUS "OpenSSL version: ${OPENSSL_VERSION}")
message(STATUS "OpenSSL include: ${OPENSSL_INCLUDE_DIR}")

# Source files
set(VAX_SOURCES
    src/gi.c
    src/sai.c
    src/verify.c
)

# Static library
add_library(vax STATIC ${VAX_SOURCES})

target_include_directories(vax
    PUBLIC 
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE 
        ${OPENSSL_INCLUDE_DIR}
)

target_link_libraries(vax
    PUBLIC OpenSSL::SSL OpenSSL::Crypto
)

# Export compile commands for clangd
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Optional: Build tests
option(BUILD_TESTS "Build test executable" ON)

if(BUILD_TESTS)
    # Test: gi
    add_executable(test_gi test/test_gi.c)
    target_link_libraries(test_gi PRIVATE vax)
    target_include_directories(test_gi PRIVATE test)
    
    # Test: sai
    add_executable(test_sai test/test_sai.c)
    target_link_libraries(test_sai PRIVATE vax)
    target_include_directories(test_sai PRIVATE test)
    
    # Test: verify
    add_executable(test_verify test/test_verify.c)
    target_link_libraries(test_verify PRIVATE vax)
    target_include_directories(test_verify PRIVATE test)
    
    # Enable testing
    enable_testing()
    add_test(NAME test_gi COMMAND test_gi)
    add_test(NAME test_sai COMMAND test_sai)
    add_test(NAME test_verify COMMAND test_verify)
    
    message(STATUS "Tests enabled - run with: ctest or ./build/test_*")
endif()

# Install rules
install(TARGETS vax
    EXPORT vaxTargets
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
)

install(FILES include/vax.h
    DESTINATION include
)

# Generate package config
install(EXPORT vaxTargets
    FILE vaxTargets.cmake
    NAMESPACE vax::
    DESTINATION lib/cmake/vax
)
