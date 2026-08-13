# Dev Context

Dev Context is a local desktop launcher for keeping personal and company development identities separate on one machine.

Instead of opening a project with:

```bash
code .
```

you open it with:

```bash
devctx .
```

Dev Context asks which context to use, then launches VS Code with isolated local state for that context:

- Separate VS Code user data
- Separate Claude Code config directory
- Separate Codex home directory
- A `DEVCTX_CONTEXT` environment marker

The goal is isolation, not credential syncing. Dev Context does not copy tokens, passwords, OAuth credentials, or provider account data between contexts.

## Install

### For Users

Users should not need Go, Node.js, npm, or Wails.

Download the latest Dev Context release from [GitHub Releases](https://github.com/Jack-WebDev/devcontext/releases), then install the artifact for your operating system.

Windows:

1. Download `devctx_<version>_windows_amd64_installer.exe`.
2. Run the installer.
3. Open Dev Context from the Start menu, or run `devctx.exe` from a terminal if the installer added it to `PATH`.

macOS:

1. Download `devctx_<version>_macos_universal.zip`.
2. Unzip it and move Dev Context to `/Applications`.
3. Open Dev Context from Applications, or add a shell shim/symlink to the bundled `devctx` executable if you want terminal access.

Linux:

1. Download `devctx_<version>_linux_amd64.tar.gz`.
2. Extract it.
3. Move the `devctx` binary somewhere on `PATH`.
4. Run `devctx` from a terminal.

Verify the install:

```bash
devctx --version
```

Dev Context expects the VS Code command-line launcher to be available as `code`.

### Package Managers

Package-manager installs are not published yet.

The intended future install experience is:

Windows:

```bash
winget install <package-id>
```

macOS:

```bash
brew install --cask <cask-name>
```

Those commands require published package manifests that point at signed release artifacts. Until those manifests exist, use GitHub Releases.

## Develop from Source

These tools are only required to build or develop Dev Context from source. End users do not need them.

Requirements:

- Go 1.25 or newer
- Node.js 22 or newer
- npm
- Wails CLI v2

Install frontend dependencies:

```bash
cd frontend
npm install
cd ..
```

Run the app in development:

```bash
npm run dev
```

Build a local production app:

```bash
npm run build
```

## Publish a Release

Releases are built by GitHub Actions when a version tag is pushed.

Create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds and uploads:

- `devctx_<version>_windows_amd64_installer.exe`
- `devctx_<version>_macos_universal.zip`
- `devctx_<version>_linux_amd64.tar.gz`
- `SHA256SUMS.txt`

After the GitHub Release is published, users can install from the release page without Go, Node.js, npm, or Wails.

Build a release binary with version metadata:

```bash
scripts/build-release.sh 1.0.0
```

## Use

Open the selector for the current directory:

```bash
devctx
```

Open a specific project:

```bash
devctx /path/to/project
```

Launch directly into a context:

```bash
devctx --context personal .
devctx --context company .
```

Use the aliases:

```bash
devctx --personal .
devctx --company .
```

Show available contexts:

```bash
devctx context list
```

Show the current project's remembered context:

```bash
devctx project show
```

Remember or remove a project binding:

```bash
devctx project bind personal
devctx project unbind
```

## Claude and Codex Setup

Having Claude or Codex subscriptions on your machine does not make a new Dev Context ready automatically.

That is intentional. Dev Context isolates provider state by pointing each launched VS Code window at context-owned directories:

- Claude: `~/.devctx/contexts/<context-id>/claude`
- Codex: `~/.devctx/contexts/<context-id>/codex`

For a new Personal or Company context, launch the context and sign in to Claude/Codex inside that VS Code session. Dev Context will report a provider as ready only when:

- The provider command exists on `PATH` (`claude` or `codex`)
- The context-owned provider directory exists
- That context-owned directory is not empty

Dev Context does not read global Claude/Codex credentials and does not copy them into a context.

## Checks

Run backend tests:

```bash
go test ./...
```

Run frontend tests and build:

```bash
cd frontend
npm run test:once
npm run build
```

Run startup benchmarks:

```bash
scripts/benchmark-startup.sh
```

Release evidence templates live in `docs/release`.

## Security

For security reports, see [SECURITY.md](SECURITY.md).

## License

Dev Context is released under the [MIT License](LICENSE).
