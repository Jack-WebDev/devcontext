<div align="center">

# DEVCTX

### Development identities, isolated.

[![Release](https://img.shields.io/github/v/release/Jack-WebDev/devcontext?style=flat-square\&label=release)](https://github.com/Jack-WebDev/devcontext/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/Jack-WebDev/devcontext/release.yml?style=flat-square\&label=build)](https://github.com/Jack-WebDev/devcontext/actions)
[![License](https://img.shields.io/github/license/Jack-WebDev/devcontext?style=flat-square)](LICENSE)

**Keep personal, company, client, and open-source development identities separate on the same machine.**

[Getting Started](#getting-started) •
[How It Works](#how-it-works) •
[Installation](#installation) •
[Roadmap](#roadmap) •
[Contributing](CONTRIBUTING.md)

</div>

---

> **Note**
>
> Dev Context is currently pre-1.0. The core workflow is usable, but commands, configuration, storage, and integrations may evolve as the project matures.

## What is it?

`devctx` is a local development identity launcher.

Instead of opening a project directly:

```bash
code .
```

open it through Dev Context:

```bash
devctx .
```

Dev Context asks **which development identity you want to use**, prepares an isolated environment for that identity, and then launches your coding tool.

```text
$ devctx .

        ┌───────────────────────────┐
        │        Choose context     │
        │                           │
        │   Personal    Company     │
        └─────────────┬─────────────┘
                      │
                      ▼
              Prepare identity
                      │
          ┌───────────┴───────────┐
          │                       │
          ▼                       ▼
       Claude                   Codex
       Personal                 Personal
          │                       │
          └───────────┬───────────┘
                      │
                      ▼
                   VS Code
```

The goal is simple:

> **Know which identity you're using before you code.**

---

## Why?

A single development machine can easily contain several identities.

You might have:

* a personal Claude subscription
* a company Claude subscription
* a personal Codex account
* a company Codex account
* different Git identities
* different GitHub accounts
* different package registries
* different cloud accounts

The problem is that development tools normally don't understand the boundary between them.

You open a personal project:

```bash
code ~/projects/my-side-project
```

but your current Claude session belongs to your employer.

Or you open company code while your personal development account is active.

Dev Context puts an explicit identity boundary **before the project launches**.

```text
Personal project
      │
      ▼
 Personal context
      │
 ┌────┴────┐
 ▼         ▼
Claude    Codex
Personal  Personal
```

and separately:

```text
Company project
      │
      ▼
 Company context
      │
 ┌────┴────┐
 ▼         ▼
Claude    Codex
Company   Company
```

Both can run at the same time without sharing provider state.

---

## Getting Started

Install Dev Context, then run:

```bash
devctx .
```

Choose a context:

```text
Personal
Company
```

Dev Context prepares that context and launches the project.

You can also launch a specific identity directly:

```bash
devctx --personal .
```

or:

```bash
devctx --company .
```

Equivalent long form:

```bash
devctx --context personal .
```

```bash
devctx --context company .
```

You can also open another project:

```bash
devctx /path/to/project
```

---

## How It Works

The central concept in Dev Context is the **context**.

A context represents a development identity.

Examples:

```text
Personal
Company
Client A
Open Source
```

A context can contain isolated configuration and state for the development tools associated with that identity.

```text
Context
│
├── Coding tool
│
├── Providers
│
├── Environment
│
└── Project
```

For example:

```text
Personal
│
├── VS Code
├── Claude Code
└── Codex
```

and:

```text
Company
│
├── VS Code
├── Claude Code
└── Codex
```

Dev Context launches each environment independently.

The context is the root concept.

VS Code, Claude Code, Codex, and future integrations are adapters that operate inside that context.

---

## Features

### Context-based launching

Launch any project through a development identity:

```bash
devctx .
```

### Multiple identities at the same time

A Personal project and Company project can run simultaneously without sharing the same provider environment.

### Claude Code isolation

Claude Code state can be isolated per context.

```text
Personal
└── Claude → Personal

Company
└── Claude → Company
```

### Codex isolation

Codex state can also be separated:

```text
Personal
└── Codex → Personal

Company
└── Codex → Company
```

### Project bindings

Projects can remember their normal identity.

```bash
devctx project bind personal
```

Now Dev Context knows that the project normally belongs to your Personal context.

Inspect the binding:

```bash
devctx project show
```

Remove it:

```bash
devctx project unbind
```

### Direct launching

Skip the selector when you already know the context:

```bash
devctx --personal .
```

### Local-first

Dev Context stores context information locally.

There is no Dev Context cloud account required to separate your local development identities.

---

## Supported Tools

### Coding tools

| Tool           | Status      |
| -------------- | ----------- |
| VS Code        | ✅ Supported |
| Cursor         | 🚧 Planned  |
| Windsurf       | 🚧 Planned  |
| JetBrains IDEs | 🗺️ Planned |
| Zed            | 🗺️ Planned |
| Neovim         | 🗺️ Planned |

### Providers

| Provider          | Status      |
| ----------------- | ----------- |
| Claude Code       | ✅ Supported |
| Codex             | ✅ Supported |
| Git identities    | 🗺️ Planned |
| GitHub identities | 🗺️ Planned |
| npm identities    | 🗺️ Planned |
| AWS profiles      | 🗺️ Planned |
| Azure profiles    | 🗺️ Planned |

The long-term model is intentionally not tied to VS Code or any single AI provider.

---

## Installation

Prebuilt binaries are available from:

**[GitHub Releases →](https://github.com/Jack-WebDev/devcontext/releases)**

### Windows

Download:

```text
devctx_<version>_windows_amd64_installer.exe
```

Run the installer and verify:

```bash
devctx --version
```

### macOS

Download:

```text
devctx_<version>_macos_universal.zip
```

Extract the archive and move Dev Context into:

```text
/Applications
```

Then verify:

```bash
devctx --version
```

### Linux

Download:

```text
devctx_<version>_linux_amd64.tar.gz
```

Extract:

```bash
tar -xzf devctx_<version>_linux_amd64.tar.gz
```

Move the executable onto your `PATH`:

```bash
sudo mv devctx /usr/local/bin/devctx
```

Verify:

```bash
devctx --version
```

---

## Requirements

The current coding-tool implementation requires the VS Code CLI:

```bash
code --version
```

Claude Code is optional:

```bash
claude --version
```

Codex is optional:

```bash
codex --version
```

Dev Context only manages integrations that you enable.

---

## Context Management

List available contexts:

```bash
devctx context list
```

Create a Personal context:

```bash
devctx context create personal
```

Create a Company context:

```bash
devctx context create company
```

Contexts are stored locally under the Dev Context home directory.

Conceptually:

```text
~/.devctx/
└── contexts/
    ├── personal/
    │   ├── context.toml
    │   └── providers/
    │       ├── claude/
    │       └── codex/
    │
    └── company/
        ├── context.toml
        └── providers/
            ├── claude/
            └── codex/
```

The exact storage structure may change before 1.0.

---

## Security

Dev Context creates **local identity boundaries**.

It is not a secrets manager.

Dev Context does not intentionally:

* parse credential secrets
* display credential secrets
* upload credential files
* synchronize credentials between contexts
* overwrite existing context credentials during normal launches

Provider credential files are treated as opaque provider-owned data.

Authentication remains the responsibility of the provider.

Dev Context controls which local environment that provider runs inside.

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

---

## Architecture

Dev Context follows one core rule:

```text
Context
   │
   ├── Coding tool adapter
   ├── Provider adapters
   ├── Environment
   └── Launch
```

Not:

```text
VS Code → Context
Claude  → Context
Codex   → Context
```

Integrations are replaceable.

The development identity is the product.

---

## Roadmap

### Identity confidence

Before a project launches, Dev Context should be able to clearly answer:

```text
Context      Personal

Project      ~/projects/devcontext
Editor       VS Code
Claude       Personal ✓
Codex        Personal ✓

Status       Ready to launch
```

Eventually additional identity checks may include:

```text
Git          jack@example.com ✓
GitHub       Jack-WebDev ✓
npm          jack ✓
AWS          personal ✓
```

### Coding tool adapters

Planned:

```text
VS Code
   ↓
Cursor
   ↓
Windsurf
   ↓
JetBrains
   ↓
Zed
   ↓
Neovim
```

### Additional identity providers

Potential integrations include:

* Git
* GitHub
* npm
* AWS
* Azure
* additional AI development tools

The scope remains development identity separation.

Dev Context is not intended to become a runtime manager, task runner, package manager, or general secrets manager.

---

## Building From Source

Requirements:

* Go 1.25+
* Node.js 22+
* npm
* Wails CLI v2
* platform-specific Wails dependencies

Clone:

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

Run:

```bash
npm run dev
```

Build:

```bash
npm run build
```

Run Go tests:

```bash
go test ./...
```

Run frontend tests:

```bash
cd frontend
npm run test:once
```

---

## Contributing

Contributions are welcome.

Bug reports, feature requests, documentation improvements, platform fixes, and new integrations are all useful.

Before making a substantial architectural change, open an issue first so the approach can be discussed.

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

Dev Context is available under the [MIT License](LICENSE).

<div align="center">

**Know which identity you're using before you code.**

</div>
