APP := fyne-template

.PHONY: build run test lint fmt vet linux-deps package artifact artifact-name clean

build:
	go build -o $(APP) .

run: build
	./$(APP)

test:
	go test -short ./...

# golangci-lint runs as a go.mod tool: the version lives in go.mod's tool
# directive, so Dependabot bumps it like any other dependency. CI runs this
# same target (.github/workflows/lint.yml).
lint:
	go tool golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# linux-deps installs the Fyne cgo build dependencies (Linux only). One
# definition shared by ci.yml and lint.yml.
linux-deps:
	sudo apt-get update
	sudo apt-get install -y \
		build-essential \
		libgl1-mesa-dev \
		xorg-dev \
		libwayland-dev \
		libxkbcommon-dev

# "make package" packages for the current OS; "make package PKGOS=windows"
# (or linux/darwin) cross-packages. -release bakes the release icon/metadata.
# PKGOS, not OS: Windows sets OS=Windows_NT in every process environment.
# fyne runs as a go.mod tool (fyne.io/tools) — Dependabot keeps it current.
package:
	go tool fyne package -release $(if $(PKGOS),-os $(PKGOS))

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
