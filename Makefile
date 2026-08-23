BINARY := shellector
BIN_DIR := bin
PREFIX := /bin

.PHONY: all build test install clean

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/shellector

test:
	go test ./...

install: build
	install -Dm755 $(BIN_DIR)/$(BINARY) $(DESTDIR)$(PREFIX)/$(BINARY)

clean:
	rm -rf $(BIN_DIR)
