APP := taskigt

ifeq ($(OS),Windows_NT)
  BINARY     := bin/$(APP).exe
  MKDIR_BIN  := if not exist bin mkdir bin
  RM_BIN     := if exist bin rmdir /s /q bin
  BUILD_TIME := v0.1_$(shell powershell -NoProfile -Command "[DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm')")
else
  BINARY     := bin/$(APP)
  MKDIR_BIN  := mkdir -p bin
  RM_BIN     := rm -rf bin
  BUILD_TIME := v0.1_$(shell date -u +%Y-%m-%dT%H:%M)
endif

FOTOGRAF := G:/fotograf/fotograf.exe

.PHONY: run test build install clean screenshot

run:
	go run ./cmd/$(APP)

test:
	go test ./...

build:
	$(MKDIR_BIN)
	go build -ldflags "-X main.BuildVersion=$(BUILD_TIME)" -o $(BINARY) ./cmd/$(APP)

install:
	go install -ldflags "-X main.BuildVersion=$(BUILD_TIME)" ./cmd/$(APP)

screenshot:
	$(FOTOGRAF) -cmd bin/$(APP).exe -args "-data,demo.json" -o screenshots/main_screen.png -wait 3s -idle 500ms -cols 80 -rows 25 -fontsize 10 -scale 3

clean:
	$(RM_BIN)
