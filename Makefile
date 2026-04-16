APP := taskigt

ifeq ($(OS),Windows_NT)
  BINARY := bin/$(APP).exe
  INSTALL_BINARY := ~/bin/$(APP).exe
else
  BINARY := bin/$(APP)
  INSTALL_BINARY := ~/bin/$(APP)
endif

.PHONY: run test build clean

run:
	go run ./cmd/$(APP)

test:
	go test ./...

BUILD_TIME := v0.1_$(shell date -u +%Y-%m-%dT%H:%M)

build:
	mkdir -p bin
	go build -ldflags "-X main.BuildVersion=$(BUILD_TIME)" -o $(BINARY) ./cmd/$(APP)

install: build
	mkdir -p ~/bin
	cp $(BINARY) $(INSTALL_BINARY)

clean:
	rm -rf bin
