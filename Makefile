APP      := fyne-template
FYNE     := fyne.io/fyne/v2/cmd/fyne@v2.8.2
GOLANGCI := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0

.PHONY: build run test lint fmt vet package artifact artifact-name clean

build:
	go build -o $(APP) .

run: build
	./$(APP)

test:
	go test -short ./...

# go run keeps contributors off a separate golangci-lint install step; the
# version pinned here is the one CI runs (lint.yml invokes this target).
lint:
	go run $(GOLANGCI) run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# "make package" packages for the current OS; "make package OS=windows"
# (or linux/darwin) cross-packages. -release bakes the release icon/metadata.
package:
	go install $(FYNE)
	fyne package -release $(if $(OS),-os $(OS))

# artifact-name prints the release zip the current OS produces. CI uses
# this for the uploaded artifact name, so naming has a single source.
artifact-name:
	@case "$$(go env GOOS)" in \
	windows) echo "$(APP)-windows-$$(go env GOARCH).exe.zip" ;; \
	darwin)  echo "$(APP)-macos.app.zip" ;; \
	linux)   echo "$(APP)-linux-$$(go env GOARCH).zip" ;; \
	*) echo "unsupported GOOS $$(go env GOOS)" >&2; exit 1 ;; \
	esac

# artifact packages the current OS and produces the release zip layouts
# the in-app updater expects (pinned by TestStagePayloadPublishedLayouts):
# linux re-packs fyne's tar.xz as a plain zip (the updater needs only
# archive/zip), windows zips the GUI-subsystem exe, darwin zips the .app.
artifact: package
	@set -eu; \
	name="$$($(MAKE) -s artifact-name)"; \
	case "$$(go env GOOS)" in \
	linux) \
	  rm -rf unpack; mkdir -p unpack; \
	  tar -xJf $(APP)*.tar.xz -C unpack; \
	  (cd unpack && zip -r "../$$name" $(APP)); \
	  ;; \
	windows) \
	  7z a -tzip "$$name" $(APP).exe; \
	  ;; \
	darwin) \
	  zip -ry "$$name" $(APP).app; \
	  ;; \
	esac; \
	echo "built $$name"

clean:
	rm -rf $(APP) $(APP).exe $(APP).app $(APP)*.tar.xz $(APP)-macos.app.zip \
	  $(APP)-linux-*.zip $(APP)-windows-*.exe.zip unpack upload
