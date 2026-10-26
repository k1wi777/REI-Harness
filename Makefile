# REI Harness — CLI `rei`
#
# Comandos:
#   make build     Compila el binario en bin/rei
#   make test      Ejecuta los tests del harness
#   make vet       Ejecuta go vet
#   make fmt       Formatea el código
#   make install   Instala el binario en GOBIN (requiere estar en PATH)
#   make cross     Compila binarios para linux/darwin/windows
#   make clean     Limpia bin/ y dist/

GO ?= go
BINARY := rei
BIN_DIR := bin
DIST_DIR := dist

# Versión inyectada en build time. `make build VERSION=v0.1.0` sella el binario.
VERSION ?= dev
PKG     := github.com/k1wi777/REI-Harness/internal/cli
LDFLAGS := -X $(PKG).version=$(VERSION)

.PHONY: build test vet fmt install cross clean

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/rei

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

install:
	$(GO) install -ldflags "$(LDFLAGS)" ./cmd/rei

cross:
	GOOS=linux   GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/rei-linux-amd64 ./cmd/rei
	GOOS=linux   GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/rei-linux-arm64 ./cmd/rei
	GOOS=darwin  GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/rei-darwin-amd64 ./cmd/rei
	GOOS=darwin  GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/rei-darwin-arm64 ./cmd/rei
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/rei-windows-amd64.exe ./cmd/rei

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)
