# Dev Context

**Dev Context is a local development identity launcher for keeping personal and professional development environments separate on the same machine.**

Instead of opening a project directly with:

```bash
code .
```

open it through Dev Context:

```bash
devctx .
```

Dev Context asks **who you are working as**, then launches Visual Studio Code with that context's isolated tool environment.

```text
devctx
   ↓
Who am I working as?
   ↓
Personal / Company / Custom Context
   ↓
VS Code + isolated Codex/Claude configuration for that identity
```

The initial use case is preventing accidental crossover between personal and company AI subscriptions, but Dev Context is built around the broader concept of **development identity** rather than any individual tool or provider.

---

## Why Dev Context?

Developers increasingly use the same computer for personal and professional work.

That computer may simultaneously contain:

* A personal Claude subscription
* A company Claude subscription
* A personal OpenAI/Codex subscription
* A company OpenAI/Codex subscription
* Different Git identities
* Different cloud accounts

Without an explicit boundary between those environments, it becomes easy to use the wrong account in the wrong project.

For example:

```text
Personal project
    ↓
Accidentally uses company Claude account
```

or:

```text
Company project
    ↓
Accidentally uses personal Codex account
```

Dev Context introduces an explicit development identity **before the editor launches**.

Instead of asking each individual tool to manage identity switching independently, Dev Context establishes the context first and launches the development environment from there.

---

## How It Works

A **context** represents a development identity.

For example:

```text
Personal Context
├── Personal Claude configuration
├── Personal Codex configuration
└── Personal tool environment
```

and:

```text
Company Context
├── Company Claude configuration
├── Company Codex configuration
└── Company tool environment
```

Each context owns its own provider state.

Dev Context currently isolates:

* Claude Code configuration
* Codex home directory
* The `DEVCTX_CONTEXT` environment variable

Dev Context does not isolate VS Code itself. VS Code keeps using the user's
normal profile, settings, extensions, and Microsoft/GitHub sign-in state.

---

## Multiple Contexts at the Same Time

Contexts are isolated through per-window provider environment variables, so multiple identities can be active simultaneously.

For example:

```text
VS Code Window A
├── Context: Personal
├── Project: constructa
├── Claude: Personal
└── Codex: Personal
```

while another window runs:

```text
VS Code Window B
├── Context: Company
├── Project: internal-api
├── Claude: Company
└── Codex: Company
```

The two environments remain separate even though they are running on the same machine.

---

## Development Identity, Not Account Switching

Dev Context is **not fundamentally an AI account switcher**.

Claude and Codex are simply the first integrations.

The core abstraction is:

```text
Context
   ↓
Development identity
   ↓
Tools and configuration
```

This allows Dev Context to grow beyond AI tooling without changing its underlying model.

Future contexts could control resources such as:

```text
Development Context
├── VS Code
├── Claude
├── Codex
├── Git identity
├── GitHub account
├── npm identity
├── AWS profile
├── Azure subscription
└── Other developer tooling
```

The long-term goal is simple:

```text
devctx
   ↓
Who am I working as?
   ↓
Everything else follows.
```

---

## Features

* Separate personal and company development identities
* Isolated Claude Code configuration
* Isolated Codex configuration
* Multiple contexts running simultaneously
* Project-to-context bindings
* Direct context launches
* Interactive desktop context selector
* Local context state
* Versioned release builds for Windows, macOS, and Linux
* No credential copying between contexts

---

## Installation

Prebuilt releases are available from [GitHub Releases](https://github.com/Jack-WebDev/devcontext/releases).

### Windows

Download:

```text
devctx_<version>_windows_amd64_installer.exe
```

Run the installer and follow the installation prompts.

After installation:

```bash
devctx --version
```

---

### macOS

Download:

```text
devctx_<version>_macos_universal.zip
```

Extract the archive and move Dev Context into:

```text
/Applications
```

The macOS release is universal and is intended to support both Intel and Apple Silicon Macs.

---

### Linux

Download:

```text
devctx_<version>_linux_amd64.tar.gz
```

Extract it:

```bash
tar -xzf devctx_<version>_linux_amd64.tar.gz
```

Move the binary somewhere on your `PATH`:

```bash
sudo mv devctx /usr/local/bin/devctx
```

Verify the installation:

```bash
devctx --version
```

---

## Requirements

Dev Context currently launches Visual Studio Code through its command-line interface.

The `code` command must therefore be available on your `PATH`.

Verify it with:

```bash
code --version
```

Claude Code and Codex are optional.

If you want Dev Context to manage either provider, their respective commands must also be available:

```bash
claude --version
```

```bash
codex --version
```

---

## Quick Start

Open the current directory:

```bash
devctx .
```

Dev Context opens the context selector and lets you choose the development identity for that project.

You can also provide another directory:

```bash
devctx /path/to/project
```

---

## Launch a Specific Context

Launch directly into a named context:

```bash
devctx --context personal .
```

```bash
devctx --context company .
```

For the built-in Personal and Company contexts, aliases are also available:

```bash
devctx --personal .
```

```bash
devctx --company .
```

This bypasses the selector when you already know which identity you want to use.

---

## Contexts

List available contexts:

```bash
devctx context list
```

A context owns the local development state associated with that identity.

Conceptually:

```text
~/.devctx/
└── contexts/
    ├── personal/
    │   ├── ...
    │   ├── claude/
    │   └── codex/
    │
    └── company/
        ├── ...
        ├── claude/
        └── codex/
```

Provider state remains isolated inside its owning context.

---

## Project Bindings

Dev Context can remember which context a project normally uses.

Bind the current project:

```bash
devctx project bind personal
```

or:

```bash
devctx project bind company
```

Show the current project's binding:

```bash
devctx project show
```

Remove the binding:

```bash
devctx project unbind
```

Project bindings reduce the chance of selecting the wrong identity for projects you open regularly.

---

## Claude Code and Codex

When a Dev Context is created, Dev Context imports supported existing local
Claude and Codex session files from the user's global provider directories into
that new context.

Each context owns separate provider directories:

```text
~/.devctx/contexts/<context-id>/claude
```

and:

```text
~/.devctx/contexts/<context-id>/codex
```

The imported files are copied to:

```text
~/.devctx/contexts/<context-id>/claude/.credentials.json
~/.devctx/contexts/<context-id>/claude/settings.json
~/.devctx/contexts/<context-id>/codex/auth.json
```

Credential files are treated as opaque files. Dev Context does not inspect,
parse, log, display, upload, or transmit credential contents.

Import happens only while explicitly creating a context. Dev Context does not
overwrite an existing context's provider credentials during normal launch.

If a global credential file is missing, that provider directory is left empty
and the provider can authenticate normally from inside the context.

For example, after creating both default contexts:

```text
Personal Context  ->  ~/.devctx/contexts/personal/{claude,codex}
Company Context   ->  ~/.devctx/contexts/company/{claude,codex}
```

Dev Context considers a provider configured when:

* The context-owned provider directory exists
* The context-owned provider directory contains provider state

---

## Isolation Model

Dev Context separates provider state by launching VS Code with context-specific provider environment variables.

Conceptually:

```text
                    Dev Context
                         │
                ┌────────┴────────┐
                │                 │
            Personal          Company
                │                 │
          ┌─────┴─────┐     ┌─────┴─────┐
          │           │     │           │
       Claude       Codex Claude       Codex
          │           │     │           │
          └─────┬─────┘     └─────┬─────┘
              isolated providers
```

Both sides can run simultaneously.

VS Code itself uses the normal user profile in both cases.

---

## Security Model

Dev Context manages **separation**, not ongoing credential synchronization.

It does not intentionally:

* Copy passwords between contexts
* Copy OAuth tokens between contexts
* Copy Claude credentials between contexts
* Copy Codex credentials between contexts
* Merge provider configuration directories
* Re-import global provider credentials during launch
* Overwrite existing isolated provider credentials

Authentication remains the responsibility of each provider.

Dev Context provides isolated locations in which those providers can maintain their own state.

For security vulnerabilities, see [SECURITY.md](SECURITY.md).

---

## What Dev Context Does Not Do

Dev Context is not:

* A password manager
* A secrets manager
* An OAuth provider
* A credential synchronization service
* A replacement for Claude Code authentication
* A replacement for Codex authentication
* A replacement for VS Code profiles
* A cloud development environment

Its responsibility is establishing **which development identity should own a development session** and launching tools accordingly.

---

## Architecture

The product is intentionally centered around contexts rather than providers.

The dependency direction should remain conceptually:

```text
Context
   ↓
Environment
   ↓
Integrations
```

rather than:

```text
Claude
   ↓
Context
```

or:

```text
Codex
   ↓
Context
```

This distinction is important because Claude and Codex are integrations, not the foundation of the application.

A future Dev Context installation may manage:

```text
Context
├── Editor state
├── AI providers
├── Git configuration
├── Source control accounts
├── Package registry identities
├── Cloud profiles
└── Environment configuration
```

without changing what a context fundamentally represents.

---

## Development

The following dependencies are required only when developing Dev Context from source.

### Requirements

* Go 1.25 or newer
* Node.js 22 or newer
* npm
* Wails CLI v2
* Platform-specific Wails dependencies

Clone the repository:

```bash
git clone https://github.com/Jack-WebDev/devcontext.git
cd devcontext
```

Install frontend dependencies:

```bash
cd frontend
npm install
cd ..
```

Run Dev Context in development mode:

```bash
npm run dev
```

Build the application locally:

```bash
npm run build
```

---

## Testing

Run the Go test suite:

```bash
go test ./...
```

Run frontend tests:

```bash
cd frontend
npm run test:once
```

Build the frontend:

```bash
npm run build
```

Run startup benchmarks:

```bash
scripts/benchmark-startup.sh
```

---

## Releases

Official release artifacts are built through GitHub Actions.

Each release currently provides:

```text
devctx_<version>_windows_amd64_installer.exe
devctx_<version>_macos_universal.zip
devctx_<version>_linux_amd64.tar.gz
SHA256SUMS.txt
```

Checksums can be used to verify downloaded release artifacts.

See [GitHub Releases](https://github.com/Jack-WebDev/devcontext/releases) for available versions.

---

## Package Managers

Package-manager distribution is planned but not currently published.

Future distribution may include:

* WinGet
* Homebrew

Until then, install Dev Context from [GitHub Releases](https://github.com/Jack-WebDev/devcontext/releases).

---

## Project Status

Dev Context is currently pre-1.0.

The core context model is usable, but commands, configuration formats, integration behavior, and internal APIs may change while the project evolves.

If you depend on Dev Context in an important workflow, review release notes before upgrading between pre-1.0 versions.

---

## Roadmap

Dev Context begins with VS Code, Claude Code, and Codex, but the context model is intended to support additional development identity resources.

Potential future integrations include:

* Git identities
* GitHub accounts
* npm identities
* AWS profiles
* Azure subscriptions
* Additional editors
* Additional AI development tools
* Context-specific environment configuration
* Improved project-context automation
* Package-manager distribution

The roadmap is guided by one principle:

```text
Choose the development identity first.
Configure tools second.
```

---

## Contributing

Contributions, bug reports, and feature discussions are welcome.

Before making a substantial architectural change, consider opening an issue first so the design can be discussed before implementation.

When contributing code:

1. Keep the context abstraction independent from individual providers.
2. Avoid introducing dependencies from core context logic into provider-specific integrations.
3. Add tests for new behavior.
4. Ensure existing tests pass.
5. Keep platform-specific behavior isolated where practical.

Run the main checks before submitting a pull request:

```bash
go test ./...
```

```bash
cd frontend
npm run test:once
npm run build
```

More detailed contributor documentation can be added to `CONTRIBUTING.md` as the project grows.

---

## Reporting Bugs

When reporting a bug, include where possible:

* Dev Context version
* Operating system
* Architecture
* VS Code version
* Relevant provider, if applicable
* Command that was executed
* Expected behavior
* Actual behavior
* Relevant logs or error messages

Please do not include passwords, tokens, OAuth credentials, or other secrets in issues.

---

## Security

Security issues should not be reported through public GitHub issues.

See [SECURITY.md](SECURITY.md) for vulnerability reporting instructions.

---

## License

Dev Context is released under the [MIT License](LICENSE).
