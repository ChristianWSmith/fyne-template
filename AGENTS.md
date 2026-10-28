# AGENTS.md

Guide for AI agents (and humans) working in this repository. This is a Fyne
v2 desktop-app template: single-package `main`, GitHub Releases distribution,
in-app auto-updater.

## Commands

```sh
make build          # go build -o fyne-template .
make run            # build + run
make test           # go test -short ./...   (CI's test command)
make lint           # golangci-lint via go tool; version lives in go.mod
make fmt            # gofmt -w .
make vet            # go vet ./...
make package        # fyne package -release (current OS); PKGOS=linux|windows|darwin to cross
make artifact       # package + produce the release zip for the current OS
make artifact-name  # print that zip's name (CI uses it for the artifact name)
```

Full (non-short) tests include one network smoke test (live GitHub release
query) — `go test ./...` without `-short` needs internet (and skips gracefully
until the repo has published its first release).

## Layout

Flat single-package layout; everything is `package main` in the repo root:

| File | Role |
|---|---|
| `main.go` | app entry: window, menus, startup update check, update dialogs |
| `prefs.go` | preference keys + helpers (update opt-out, window size) |
| `version.go` | **stamp-only**: `var version = "dev"` — CI overwrites the whole file on release builds; never put logic here |
| `semver.go` | strict `X.Y.Z` parse/compare, `updateAvailable` |
| `update.go` | release check, download, zip extract (zip-slip safe), payload staging, `cleanupOlds` |
| `update_{linux,windows,darwin,other}.go` | build-tagged `replaceAndRelaunch` per platform |
| `FyneApp.toml` | app Name/ID/Icon/Version — `fyne package` reads this |
| `icon.png` | 512×512; embedded at compile time (`//go:embed`) for the window icon, referenced by `FyneApp.toml` for packaging |

## Release pipeline

- CI (`.github/workflows/ci.yml`) builds only on `workflow_dispatch` (release)
  and `pull_request`. Pushes never build. The dispatch input `version`
  (`X.Y.Z`) is stamped over `version.go` before compiling. Build, test, and
  packaging steps call the Makefile (`make vet`, `make test`, `make artifact`)
  so CI and local flows are the same commands; the uploaded artifact name
  comes from `make artifact-name`.
- Artifacts: `fyne-template-linux-amd64.zip` (tar.xz tree re-zipped),
  `fyne-template-windows-amd64.exe.zip`, `fyne-template-macos.app.zip`.
  These exact layouts are pinned by `TestStagePayloadPublishedLayouts` —
  change the Makefile `artifact` target and `update.go` staging in the same
  commit. Lint CI (`.github/workflows/lint.yml`) also runs `make lint`.
- Tags are plain semver `X.Y.Z` (no `v`), created only after the dispatch
  build is green. The `release` skill (`.opencode/skills/release/SKILL.md`)
  is the canonical cut-a-release procedure.
- The updater queries `https://api.github.com/repos/<updateRepo>/releases/latest`
  (`updateRepo` const in `update.go`). `version = "dev"` never update-checks.

## Conventions

- **Formatting/lint:** golangci-lint (`.golangci.yml`) + `go vet`; no other
  style tooling. Keep `make lint` green.
- **Tests:** `-short` must skip anything network-bound (`testing.Short()`).
- **UI threading:** never touch widgets off the fyne thread; marshal with
  `fyne.Do(func(){ ... })`. Check the Fyne documentation for threading
  rules when building complex UI.
- **Prefs:** keys live in `prefs.go`; "Never" opt-outs are written
  immediately (a crash must not resurrect a dismissed prompt), not batched
  to app close.
- **version.go stays stamp-only.** The CI stamp step replaces the file
  wholesale.
- **`make package` rewrites `FyneApp.toml`** — the fyne CLI re-serializes
  the file and increments `Build` on every run. Local-only noise (CI's
  copy is ephemeral); discard with `git checkout -- FyneApp.toml` when
  you don't want the bump committed.

## Renaming the template

The identity tokens are `fyne-template` (app name, asset prefix, module tail)
and `ChristianWSmith` (owner). Grep both and update: `go.mod`,
`FyneApp.toml` (Name + ID), `update.go` (`appName`, `updateRepo`), `main.go`
(`app.NewWithID`), `Makefile` (`APP`), `.gitignore` (binary lines), plus
`icon.png`. `appName` must equal `FyneApp.toml` `Name` — the
updater derives asset and binary names from it.
