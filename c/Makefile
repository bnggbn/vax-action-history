# VAX C Library Makefile

CC = gcc
CFLAGS = -Wall -Wextra -std=c11 -O2 -Iinclude
LDFLAGS = -lssl -lcrypto

# Source files
SRCS = src/gi.c src/sai.c src/verify.c
OBJS = $(SRCS:.c=.o)

# Output library
TARGET = libvax.a

# Test executable
TEST = vax_test

.PHONY: all clean test

all: $(TARGET)

# Build static library
$(TARGET): $(OBJS)
	ar rcs $@ $^
	@echo "Built: $@"

# Compile object files
%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

# Build test executable (if you have a test.c)
test: $(OBJS) test.c
	$(CC) $(CFLAGS) test.c $(OBJS) $(LDFLAGS) -o $(TEST)
	./$(TEST)

# Clean build artifacts
clean:
	rm -f $(OBJS) $(TARGET) $(TEST)
	@echo "Cleaned build artifacts"

# Install (optional)
install: $(TARGET)
	cp $(TARGET) /usr/local/lib/
	cp include/vax.h /usr/local/include/
	@echo "Installed to /usr/local"
