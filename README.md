<div align="center">

# Dev Context

### Development identities, isolated.

Keep personal, company, client, and open-source development environments separate on the same machine.

<p>
  <img src="https://img.shields.io/github/v/release/Jack-WebDev/devcontext?style=flat-square&label=release&color=167246" alt="Latest release" />
  <img src="https://img.shields.io/github/actions/workflow/status/Jack-WebDev/devcontext/release.yml?style=flat-square&label=release%20build" alt="Release build" />
  <img src="https://img.shields.io/github/license/Jack-WebDev/devcontext?style=flat-square" alt="MIT license" />
</p>

[Get started](#get-started) · [Install](#installation) · [CLI reference](#cli-reference) · [Security](#security) · [Contribute](CONTRIBUTING.md)

</div>

## What Dev Context does

Dev Context is a local desktop application and CLI for launching a project with the right development identity.

Instead of opening a project directly:

```bash
code .
```

open it through Dev Context:

```bash
devctx .
```

Dev Context resolves the project, lets you choose an identity, checks that identity's local setup, and launches the configured coding tool with isolated provider state.

```text
Project
   │
   ▼
Context: Personal, Company, Client, …
   │
   ├── Coding tool configuration
   ├── Claude Code state
   ├── Codex state
   └── Context environment
```

Personal and company workspaces can run at the same time without sharing their Claude Code or Codex sessions.

## Why it exists

A development machine often carries several identities at once: personal and company AI subscriptions, client accounts, Git identities, registry credentials, and cloud profiles. Most coding tools do not enforce a boundary between them.

Dev Context puts that boundary before launch. Every project opens through an explicit context, so you can see which identity will be used before any coding tool starts.

## Highlights

- Desktop management for contexts, projects, running workspaces, launch history, settings, and diagnostics.
- Interactive context selection with preflight readiness and identity checks.
- Direct, scriptable launches with `--context`.
- Project bindings that remember a project's normal context.
- Concurrent workspaces using different contexts.
- Isolated Claude Code and Codex provider storage.
- Context import, export, duplication, archiving, and recovery workflows.
- Local activity history and actionable system diagnostics.
- Local-first operation with no Dev Context cloud account.
- Native packages for Windows, macOS, and Linux.

## Get started

1. [Install the release for your platform](#installation).
2. Confirm that the VS Code CLI is available:

   ```bash
   code --version
   ```

3. Open a terminal in a project and run:

   ```bash
   devctx .
   ```

4. Create or choose a context, review its readiness, and launch the project.

To open a different directory:

```bash
devctx /path/to/project
```

Once a context is configured, launch it directly:

```bash
devctx --context personal .
```

The built-in Personal and Company contexts also have shortcuts:

```bash
devctx --personal .
devctx --company .
```

## Installation

Download the current release and its checksum from [GitHub Releases](https://github.com/Jack-WebDev/devcontext/releases).

### Windows

Download and run:

```text
devctx_<version>_windows_amd64_installer.exe
```

The installer adds `devctx` to `PATH`. Open a new terminal after installation, then verify:

```powershell
devctx --version
```

### macOS

Download and extract:

```text
devctx_<version>_macos_universal.zip
```

Install the application and included CLI symlink:

```bash
sudo mv devctx/devctx.app /Applications/
sudo mkdir -p /usr/local/bin
sudo mv devctx/devctx /usr/local/bin/devctx
devctx --version
```

The release is a universal application for Apple silicon and Intel Macs.

### Linux

Download and extract:

```bash
tar -xzf devctx_<version>_linux_amd64.tar.gz
sudo mv devctx /usr/local/bin/devctx
devctx --version
```

## Requirements

Dev Context uses VS Code as its built-in coding tool and expects the `code` command to be available on `PATH`.

Claude Code and Codex are optional. Dev Context only manages provider integrations enabled for a context.

```bash
code --version
claude --version  # optional
codex --version   # optional
```

## Core concepts

### Contexts

A context is a named development identity such as Personal, Company, or Client A. It owns the coding-tool configuration, provider storage, and environment used for its launches.

Contexts can be created in the desktop app or from the CLI:

```bash
devctx context create personal
devctx context list
```

### Project bindings

A project binding records the context a project normally uses. Interactive launches can remember this choice, or it can be managed explicitly:

```bash
devctx project bind personal
devctx project show
devctx project unbind
```

Bindings protect against accidental identity changes. If a direct launch requests a context different from the remembered context, Dev Context stops before launching.

For deliberate non-interactive automation, acknowledge the override explicitly:

```bash
devctx --context company --allow-context-mismatch .
```

### Readiness and diagnostics

Before launch, Dev Context checks the selected context, project binding, coding tool, provider storage, and isolation configuration. Blocking problems stop the launch; recoverable conditions are shown with guidance.

The desktop app's System Health screen exposes application-wide and context-specific diagnostics without displaying credential contents.

## CLI reference

| Command | Purpose |
| --- | --- |
| `devctx [path]` | Open the interactive context selector for a project. |
| `devctx --context <id> [path]` | Launch directly with a context. |
| `devctx --personal [path]` | Launch with the built-in Personal context. |
| `devctx --company [path]` | Launch with the built-in Company context. |
| `devctx context list` | List contexts. |
| `devctx context create <id>` | Create a context. |
| `devctx project show` | Show the current project's binding. |
| `devctx project bind <id>` | Bind the current project to a context. |
| `devctx project unbind` | Remove the current project's binding. |
| `devctx --version` | Print version and build information. |

Add `--debug` to a launch command when collecting troubleshooting information.

## Supported integrations

| Category | Integration | Status |
| --- | --- | :---: |
| Coding tool | VS Code | Supported |
| Provider | Claude Code | Supported |
| Provider | Codex | Supported |

The integration model is context-first: coding tools and providers are replaceable adapters, while the development identity remains the root concept.

## Local data

Dev Context stores configuration and context state under `~/.devctx`:

```text
~/.devctx/
├── config.toml
├── contexts/
│   └── <context-id>/
│       ├── context.toml
│       ├── tools/
│       └── providers/
├── projects.toml
├── recents.toml
├── running.toml
└── logs/
```

Project files remain in their original directories. Forgetting a project removes Dev Context's binding, recent-launch, and activity records; it never deletes the project directory.

## Security

Dev Context creates local identity boundaries. It is not a password manager, secrets manager, sandbox, or replacement for your operating system's access controls.

Normal launches do not copy, move, replace, or synchronize global provider credentials. When you explicitly import a detected provider session while creating a context, Dev Context copies that provider's local credential file into the new context. Provider adapters may read allowlisted identity metadata for local display, but credential contents are never displayed, logged, uploaded, or included in exports.

Dev Context does not require a hosted account and does not intentionally upload local configuration or credential files.

See [SECURITY.md](SECURITY.md) for the vulnerability-reporting process. Never include tokens, credentials, or sensitive company information in a public issue.

## Build from source

Requirements:

- Go 1.25 or later
- Node.js 22 or later
- npm
- Wails CLI v2
- [platform-specific Wails dependencies](https://wails.io/docs/gettingstarted/installation)

```bash
git clone https://github.com/Jack-WebDev/devcontext.git
cd devcontext

cd frontend
npm install
cd ..

npm run dev
```

Run the release checks before submitting a change:

```bash
go test ./...
npm --prefix frontend run build
npm --prefix frontend run test:once
```

Build the desktop application with:

```bash
npm run build
```

On Linux, the repository scripts use the WebKitGTK 4.1 build tag.

## Contributing and support

Bug reports, documentation fixes, platform improvements, and new integrations are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

When reporting a problem, include:

- Dev Context version (`devctx --version`)
- operating system and architecture
- command or desktop workflow used
- expected and actual behavior
- sanitized diagnostic output, if relevant

Do not include credentials or private project data.

## License

Dev Context is available under the [MIT License](LICENSE).

<div align="center">

**Know which identity you're using before you code.**

</div>
