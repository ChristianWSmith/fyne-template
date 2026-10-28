# fyne-template

A batteries-included starter template for [Fyne](https://fyne.io) desktop
apps in Go. It ships a working hello-world window plus the plumbing that's
painful to bolt on later: a multi-platform CI build, a semver release
pipeline driven by an agent skill, and a GitHub-Releases-powered in-app
auto-updater with preferences.

## What's included

- **Minimal Fyne app** — single-window hello app, embedded `icon.png`,
  `FyneApp.toml` app metadata
- **In-app auto-updater** (`update.go`, `semver.go`) — checks GitHub
  Releases on launch, offers **Update / Later / Never**, downloads the
  platform zip, swaps the running binary (or `.app` bundle), and relaunches.
  "Never" persists via Fyne preferences; Help → Check for Updates bypasses it
- **Preferences persistence** (`prefs.go`) — the update opt-out plus window-size
  restore, as a minimal example of `fyne.Preferences`
- **CI** (`.github/workflows/ci.yml`) — PR validation plus on-demand
  `workflow_dispatch` builds that stamp a version into the binary and produce
  Linux, Windows, and macOS release zips via the same `make` targets used
  locally
- **Lint** (`.github/workflows/lint.yml`) — `make lint` (golangci-lint) on
  every push/PR
- **Release skill** (`.opencode/skills/release/SKILL.md`) — agent-driven
  pipeline: checks → commit → dispatch CI → tag-after-green → publish GitHub
  release with the three artifacts attached
- **Branding skill** (`.opencode/skills/branding/SKILL.md`) — agent-driven
  one-shot conversion of this template into your app repo: gathers the app
  identity, renames every token, rewrites this README for your users, updates
  AGENTS.md and the release skill, gates on a green build, then deletes itself
- **Makefile** — `build`, `run`, `test`, `lint`, `fmt`, `vet`, `package`, `clean`
- **Dependabot** — weekly Go module and GitHub Actions update PRs

## Using this template

The easiest path: run the **branding skill** (`.opencode/skills/branding`)
from an agent session — it asks for your app name, GitHub repo, ID, and
description, performs every rename below, rewrites this README as your app's
user-facing landing page, verifies the result, and removes itself.

If you'd rather rename by hand, the identity tokens are `fyne-template` and
`ChristianWSmith`:

| Where | What |
|---|---|
| `go.mod` | module path |
| `FyneApp.toml` | `Name`, `ID` |
| `update.go` | `appName`, `updateRepo` consts |
| `main.go` | `app.NewWithID(...)`, window title |
| `Makefile` | `APP` |
| `.gitignore` | binary name lines |
| `icon.png` | replace with your app icon (512×512 PNG) |

`updateRepo` must be the real `owner/repo` of your app or the updater will
query the wrong (or non-existent) GitHub project. `appName` must match the
name `fyne package` produces (`FyneApp.toml` `Name`) — the updater looks for
release assets and binaries named after it.

## Building

```sh
make build          # go build → ./fyne-template
make run            # build + run
make test           # go test -short ./...
make lint           # golangci-lint (pinned via go run)
make package        # fyne package -release for the current OS
make package PKGOS=windows # cross-package (needs mingw for cgo)
make artifact       # package + build the release zip for this OS
```

Linux needs the usual Fyne cgo deps to build:
`build-essential libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev`.

## Releasing

Releases are cut with the `release` agent skill (`.opencode/skills/release/SKILL.md`):
it gates on lint/vet/test, bumps `FyneApp.toml`, dispatches the CI matrix with
the version to stamp, tags the exact green commit with semver `X.Y.Z` (no `v`
prefix), and publishes a GitHub release with the three platform zips attached.
Installed apps pick up the new version on their next launch.

Manual path, summarized: `gh workflow run .github/workflows/ci.yml -f
version=X.Y.Z`, wait for green, tag, `gh release create X.Y.Z <zips>`.

## License

MIT — see [LICENSE](LICENSE).
