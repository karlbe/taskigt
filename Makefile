APP := taskigt

ifeq ($(OS),Windows_NT)
  BINARY         := bin/$(APP).exe
  INSTALL_BINARY := $(USERPROFILE)/bin/$(APP).exe
  MKDIR_BIN      := if not exist bin mkdir bin
  MKDIR_USERBIN  := if not exist "$(USERPROFILE)\bin" mkdir "$(USERPROFILE)\bin"
  RM_BIN         := if exist bin rmdir /s /q bin
  CP             := copy /y
  BUILD_TIME     := v0.1_$(shell powershell -NoProfile -Command "[DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm')")
else
  BINARY         := bin/$(APP)
  INSTALL_BINARY := ~/bin/$(APP)
  MKDIR_BIN      := mkdir -p bin
  MKDIR_USERBIN  := mkdir -p ~/bin
  RM_BIN         := rm -rf bin
  CP             := cp
  BUILD_TIME     := v0.1_$(shell date -u +%Y-%m-%dT%H:%M)
endif

.PHONY: run test build install clean

run:
	go run ./cmd/$(APP)

test:
	go test ./...

build:
	$(MKDIR_BIN)
	go build -ldflags "-X main.BuildVersion=$(BUILD_TIME)" -o $(BINARY) ./cmd/$(APP)

install: build
	$(MKDIR_USERBIN)
	$(CP) $(BINARY) $(INSTALL_BINARY)

clean:
	$(RM_BIN)
