---
name: branding
description: Use when converting this template repo into a real app repo — e.g. "brand this repo", "convert the template to my app", "set up the app identity", "rename the template", "this is no longer a template". Gathers the app identity (name, GitHub repo, app ID, description, feature bullets), renames every identity token across code and config, rewrites README.md as a user-facing doc, updates AGENTS.md and the release skill, gates on fmt/lint/vet/test/build, then deletes this skill. Never commits.
---

# Brand the Repo

Convert this repository from the Fyne app template into a real app repo:
adopt the app's identity everywhere, rewrite the docs for the new audience,
and remove this skill once the conversion is verified. Nothing is committed
— the user reviews and commits themselves.

Two confirmation checkpoints (identity table, README draft). Everything
else runs autonomously. The repo must stay buildable and green at every
phase; the final gate must pass before this skill deletes itself.

## Phase 0 — Gather identity (checkpoint #1)

Ask the user for all of these in one round:

1. **App name** — the binary/package name, kebab-case. Must match
   `^[a-z0-9][a-z0-9-]*$`; reject uppercase/spaces (it becomes asset
   names, file names, and a module path segment).
2. **GitHub repo** — `owner/name`. Verify it exists:
   ```sh
   gh repo view <owner>/<name>
   ```
   If gh can't reach it (not created/pushed yet), warn the user and ask
   whether to continue anyway.
3. **App ID** — reverse-DNS, default `com.<owner-lowercased>.<app-name>`.
   Must match `^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`.
4. **One-line description** — what the app does, for AGENTS.md and the
   README tagline.
5. **Feature bullets** — 3–6 short, user-facing bullets for the README.

State up front: the icon and LICENSE are left alone, and **no commit will
be made**. Show the gathered values as a table and **ask the user to
confirm** before touching any file. A wrong identity is expensive to
unwind; don't start renames on an unconfirmed table.

Record the confirmed values as you go: `<name>`, `<owner>/<repo>`, `<id>`.

## Phase 1 — Code and config renames

Make exactly these edits (tests need no changes — they read `appName`):

1. `go.mod` — module path → `github.com/<owner>/<repo>`
2. `update.go` — the two identity consts:
   - `appName = "<name>"`
   - `updateRepo = "<owner>/<repo>"`
3. `main.go` — `app.NewWithID("<id>")` and `a.NewWindow("<name>")`
4. `FyneApp.toml` — `Name = "<name>"`, `ID = "<id>"` (leave Version/Build)
5. `Makefile` — `APP := <name>` (every other target derives from it)
6. `.gitignore` — the three binary lines → `/<name>`, `/<name>.exe`,
   `<name>.app/`
7. `update_test.go` — reword the "fresh clone of the template" comment
   (the 404-skip message) so it doesn't say "template"

**Consistency rule:** `appName` (update.go), `FyneApp.toml` `Name`, and
`Makefile` `APP` must all be the same string — the updater derives asset
and binary names from it, `fyne package` derives the binary/tar.xz name
from `FyneApp.toml`, and the `artifact` target globs `$(APP)*.tar.xz`.
A mismatch ships an updater that can't find its own assets.

**App ID rule:** `FyneApp.toml` `ID` and `app.NewWithID` must be the same
string — the packaged metadata and the runtime prefs store must agree.

Verify Phase 1 before moving on — these greps must print nothing:

```sh
grep -rn "fyne-template\|ChristianWSmith" \
  --include="*.go" --include="*.toml" --include="*.yml" \
  --include="Makefile" --include="go.mod" --include=".gitignore" .
grep -rn "template" --include="*.go" .
```

## Phase 2 — Update the release skill

`.opencode/skills/release/SKILL.md`: replace every occurrence of
`fyne-template` with `<name>`. Every occurrence in that file is app
identity (artifact names, `/tmp/fyne-template-release` scratch dirs,
linux-tree unzip examples), so a plain token replace is safe:

```sh
sed -i 's/fyne-template/<name>/g' .opencode/skills/release/SKILL.md
```

(Substitute the real name for `<name>`; don't run it literally.)

## Phase 3 — Update AGENTS.md (targeted edits)

1. **Intro** — change "This is a Fyne v2 desktop-app template:" to
   "This is a Fyne v2 desktop app: *<description>*," keeping the rest of
   the sentence (single-package `main`, GitHub Releases distribution,
   in-app auto-updater).
2. **Commands** — `make build` comment → `go build -o <name> .`
3. **Release pipeline** — the three artifact names → `<name>-linux-amd64.zip`,
   `<name>-windows-amd64.exe.zip`, `<name>-macos.app.zip`
4. **Delete the entire "## Renaming the template" section** (this skill
   just performed it).
5. Leave the layout table, conventions, and gotchas alone unless they
   still name the template — if a `fyne-template` reference remains after
   steps 1–4, fix it.

## Phase 4 — Rewrite README.md (checkpoint #2)

The README becomes **user-facing**: written for people who downloaded the
app, not for people hacking on it. Draft the full file from this skeleton,
filling in the Phase 0 values, then **show the draft to the user and ask
for confirmation** before writing it:

```markdown
# <App Name>

<One-line description.>

## Features

- <feature bullet>
- ...

## Download

Grab the latest release for your platform from the
[Releases page](https://github.com/<owner>/<repo>/releases/latest):

- `<name>-linux-amd64.zip`
- `<name>-windows-amd64.exe.zip`
- `<name>-macos.app.zip`

The app checks GitHub Releases on launch and offers to install updates
itself.

## Building from source

Requires Go (see `go.mod`) and, on Linux, the Fyne cgo dependencies:

    sudo apt-get install build-essential libgl1-mesa-dev xorg-dev \
      libwayland-dev libxkbcommon-dev

    make build    # build
    make run      # build + run
    make package  # distributable for this OS

See [AGENTS.md](AGENTS.md) for development details.

## License

MIT — see [LICENSE](LICENSE).
```

Rules for the draft:

- **No** mention of templates, this branding skill, CI internals, the
  release skill, or agent workflows — end users don't care.
- Feature bullets are plain language; no implementation details.
- Keep it short — this is a landing page, not a manual.

## Phase 5 — Gate

All must pass; fix failures before continuing. Never delete this skill
on a red tree.

```sh
gofmt -l .        # must print nothing
make lint
make vet
make test
make build
make clean        # the build just produced a binary whose embedded source
                  # paths can still contain the old directory name — don't
                  # let it pollute the grep pass below
```

Then confirm no identity or skill references survive anywhere:

```sh
grep -rn "fyne-template\|ChristianWSmith" --exclude-dir=.git --exclude-dir=branding .
grep -rn "skills/branding" --exclude-dir=.git --exclude-dir=branding .
```

Both must print nothing. (`LICENSE` says "Christian Smith" — that's the
license holder, not an identity token; it stays. Ordinary prose uses of
the word "branding" — e.g. in an app description — are fine; what must
not survive is any reference to *this skill*.)

If `grep` still hits, the rename missed a spot — find it, fix it, re-run
the full gate. Do not proceed with a known leftover.

## Phase 6 — Self-delete and report

Only after a green Phase 5:

1. Delete this skill:
   ```sh
   rm -rf .opencode/skills/branding
   ```
2. Report to the user:
   - the adopted identity (name, repo, app ID)
   - files changed per phase (renames, release skill, AGENTS.md, README)
   - gate results (fmt/lint/vet/test/build all green)
   - **the changes are uncommitted** — review with `git status` /
     `git diff`, then commit when ready

If anything in Phase 5 failed and couldn't be fixed, stop before the
delete and report — the skill must remain available to retry.

## Gotchas

- **The three-way name match** (`appName` = `FyneApp.toml` `Name` =
  `Makefile` `APP`) is the single most important invariant; the whole
  update pipeline silently depends on it. Double-check it in Phase 5's
  grep pass, not just Phase 1.
- **Never skip checkpoint #1.** Renaming on a wrong or unvalidated repo
  name produces a module path and updater that point at a repo that
  doesn't exist — the app will never find its own releases.
- The README is for **end users**. If a sentence would only make sense
  to someone reading the source, cut it; development details live in
  AGENTS.md (which this skill edits, not replaces).
- This skill deletes itself — that's intentional and only happens on a
  green gate. If the run is interrupted before Phase 6, the skill file is
  still on disk and the whole run can simply be repeated (Phases are
  idempotent: re-grepping and re-replacing an already-branded repo is a
  no-op, except the README checkpoint will re-confirm the draft).
- No commit, ever: the user asked to review branding changes themselves.
  Don't stage, commit, or push — not even "just the renames".
