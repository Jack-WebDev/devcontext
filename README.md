<div align="center">

# DEVCTX

### *Development identities, isolated.*

<p>
  <img src="https://img.shields.io/github/v/release/Jack-WebDev/devcontext?style=for-the-badge&label=release&color=6366F1&labelColor=1a1a2e" alt="Release" />
  <img src="https://img.shields.io/github/actions/workflow/status/Jack-WebDev/devcontext/release.yml?style=for-the-badge&label=build&color=22C55E&labelColor=1a1a2e" alt="Build" />
  <img src="https://img.shields.io/github/license/Jack-WebDev/devcontext?style=for-the-badge&color=F59E0B&labelColor=1a1a2e" alt="License" />
</p>

<p>
  <img src="https://img.shields.io/badge/Windows-supported-0078D6?style=flat-square&logo=windows&logoColor=white" alt="Windows" />
  <img src="https://img.shields.io/badge/macOS-supported-000000?style=flat-square&logo=apple&logoColor=white" alt="macOS" />
  <img src="https://img.shields.io/badge/Linux-supported-FCC624?style=flat-square&logo=linux&logoColor=black" alt="Linux" />
  <img src="https://img.shields.io/badge/status-pre--1.0-EF4444?style=flat-square" alt="Pre-1.0" />
</p>

**Keep personal, company, client, and open-source development identities separate on the same machine.**

<sub>One command. One prompt. Zero cross-contamination between your Claude, Codex, and editor sessions.</sub>

<br>

[**🚀 Getting Started**](#-getting-started) •
[**⚙️ How It Works**](#️-how-it-works) •
[**📦 Installation**](#-installation) •
[**🗺️ Roadmap**](#️-roadmap) •
[**🤝 Contributing**](CONTRIBUTING.md)

</div>

<br>

> [!NOTE]
> Dev Context is currently **pre-1.0**. The core workflow is usable, but commands, configuration, storage, and integrations may evolve as the project matures.

<br>

## What is it?

`devctx` is a **local development identity launcher**.

Instead of opening a project directly:
Instead of opening a project directly:

```bash
code .
```

open it through Dev Context:

```bash
devctx .
```

Dev Context asks **which development identity you want to use**, prepares an isolated environment for that identity, and then launches your coding tool.
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

<div align="center">

### Know **which identity** you're using *before* you code

</div>

<br>

## Why?

A single development machine can easily contain several identities:

| 🪪 Identity type | Example |
| --- | --- |
| AI provider | Personal & company Claude subscriptions |
| AI provider | Personal & company Codex accounts |
| Version control | Different Git identities |
| Hosting | Different GitHub accounts |
| Packages | Different package registries |
| Cloud | Different cloud accounts |

The problem is that development tools normally don't understand the boundary between them.

You open a personal project:

```bash
code ~/projects/my-side-project
```

but your current Claude session belongs to your employer. Or you open company code while your personal development account is active.

**Dev Context puts an explicit identity boundary *before* the project launches.**

<table>
<tr>
<td valign="top">

**Personal project**

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

</td>
<td valign="top">

**Company project**

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

</td>
</tr>
</table>

✅ Both can run **at the same time** without sharing provider state.

<br>

## 🚀 Getting Started

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

```bash
devctx --company .
```

Equivalent long form:

```bash
devctx --context personal .
devctx --context company .
```

Or open another project:

```bash
devctx /path/to/project
```

<br>

## ⚙️ How It Works

The central concept in Dev Context is the **context** — a development identity.

```text
Personal   Company   Client A   Open Source
```

A context can contain isolated configuration and state for the development tools associated with that identity:

```text
Context
│
├── 🖥️  Coding tool
│
├── 🔌 Providers
│
├── 🌐 Environment
│
└── 📁 Project
```

For example:

<table>
<tr>
<td valign="top">

**Personal**

```text
Personal
│
├── VS Code
├── Claude Code
└── Codex
```

</td>
<td valign="top">

**Company**

```text
Company
│
├── VS Code
├── Claude Code
└── Codex
```

</td>
</tr>
</table>

Dev Context launches each environment independently. **The context is the root concept** — VS Code, Claude Code, Codex, and future integrations are adapters that operate inside that context.

<br>

## Features

<table>
<tr>
<td width="50%" valign="top">

### 🎛️ Context-based launching

Launch any project through a development identity:

```bash
devctx .
```

### 🔀 Multiple identities at once

A Personal project and Company project can run simultaneously without sharing the same provider environment.

### 🤖 Claude isolation

```text
Personal → Claude → Personal
Company  → Claude → Company
```

</td>
<td width="50%" valign="top">

### 🧩 Codex isolation

```text
Personal → Codex → Personal
Company  → Codex → Company
```

### 📌 Project bindings

Projects can remember their normal identity:

```bash
devctx project bind personal
devctx project show
devctx project unbind
```

### ⚡ Direct launching

Skip the selector when you already know the context:

```bash
devctx --personal .
```

### 🔒 Local-first

No Dev Context cloud account required — everything stays on your machine.

</td>
</tr>
</table>

<br>

## 🧰 Supported Tools

### Coding tools

| Tool | Status |
| --- | :---: |
| VS Code | ✅ Supported |
| Cursor | 🚧 Planned |
| Windsurf | 🚧 Planned |
| JetBrains IDEs | 🗺️ Planned |
| Zed | 🗺️ Planned |
| Neovim | 🗺️ Planned |

### Providers

| Provider | Status |
| --- | :---: |
| Claude Code | ✅ Supported |
| Codex | ✅ Supported |
| Git identities | 🗺️ Planned |
| GitHub identities | 🗺️ Planned |
| npm identities | 🗺️ Planned |
| AWS profiles | 🗺️ Planned |
| Azure profiles | 🗺️ Planned |

> The long-term model is intentionally **not** tied to VS Code or any single AI provider.

<br>

## 📦 Installation

Prebuilt binaries are available from:

<div align="center">

### **[⬇️ GitHub Releases →](https://github.com/Jack-WebDev/devcontext/releases)**

</div>

<details>
<summary><b>🪟 Windows</b></summary>
<br>

Download:

```text
devctx_<version>_windows_amd64_installer.exe
```

Run the installer and verify:
Run the installer and verify:

```bash
devctx --version
```

</details>

<details>
<summary><b>🍎 macOS</b></summary>
<br>

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

</details>

<details>
<summary><b>🐧 Linux</b></summary>
<br>

Download:

```text
devctx_<version>_linux_amd64.tar.gz
```

Extract:
Extract:

```bash
tar -xzf devctx_<version>_linux_amd64.tar.gz
```

Move the executable onto your `PATH`:
Move the executable onto your `PATH`:

```bash
sudo mv devctx /usr/local/bin/devctx
```

Verify:
Verify:

```bash
devctx --version
```

</details>

<br>

## ✅ Requirements

The current coding-tool implementation requires the VS Code CLI:

```bash
code --version
```

Claude Code is **optional**:

```bash
claude --version
```

Codex is **optional**:

```bash
codex --version
```

> Dev Context only manages integrations that you enable.

<br>

## 🗂️ Context Management

```bash
# List available contexts
devctx context list

# Create a Personal context
devctx context create personal

# Create a Company context
devctx context create company
```

Contexts are stored locally under the Dev Context home directory:

```text
~/.devctx/
└── contexts/
    ├── personal/
    │   ├── context.toml
    │   └── providers/
    │       ├── claude/
    │       └── codex/
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
        ├── context.toml
        └── providers/
            ├── claude/
            └── codex/
```

> ⚠️ The exact storage structure may change before 1.0.

<br>

## 🔐 Security

Dev Context creates **local identity boundaries**. It is **not** a secrets manager.

Dev Context does not intentionally:

- ❌ parse credential secrets
- ❌ display credential secrets
- ❌ upload credential files
- ❌ synchronize credentials between contexts
- ❌ overwrite existing context credentials during normal launches

Provider credential files are treated as **opaque, provider-owned data**. Authentication remains the responsibility of the provider — Dev Context controls *which local environment* that provider runs inside.

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

<br>

## 🏗️ Architecture

Dev Context follows one core rule:

```text
Context
   │
   ├── Coding tool adapter
   ├── Provider adapters
   ├── Environment
   └── Launch
```

**Not** this — integrations are replaceable, identity is not:

```text
Coding tool → Context
Provider    → Context
Environment → Context
```

> The development identity is the product.

<br>

## 🗺️ Roadmap

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

Eventually, additional identity checks may include:

```text
Git          jack@example.com ✓
GitHub       Jack-WebDev ✓
npm          jack ✓
AWS          personal ✓
```

### 🧩 Coding tool adapters

```text
VS Code → Cursor → Windsurf → JetBrains → Zed → Neovim
```

### 🔌 Additional identity providers

Potential integrations include Git, GitHub, npm, AWS, Azure, and additional AI development tools.

Claude Code and Codex are the built-in providers today. They are provider adapters, not architectural assumptions: a new provider can contribute its own isolated storage, environment, credential handling, identity metadata, and setup guidance through the provider registry.

> The scope remains **development identity separation**. Dev Context is not intended to become a runtime manager, task runner, package manager, or general secrets manager.

<br>

## 🛠️ Building From Source

**Requirements**

- Go 1.25+
- Node.js 22+
- npm
- Wails CLI v2
- platform-specific Wails dependencies

```bash
# Clone
git clone https://github.com/Jack-WebDev/devcontext.git
cd devcontext

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run
npm run dev

# Build
npm run build

# Run Go tests
go test ./...

# Run frontend tests
cd frontend
npm run test:once
```

<br>

## 🤝 Contributing

Contributions are welcome! Bug reports, feature requests, documentation improvements, platform fixes, and new integrations are all useful.

> Before making a substantial architectural change, please open an issue first so the approach can be discussed.

See [CONTRIBUTING.md](CONTRIBUTING.md).

<br>

## 📄 License

Dev Context is available under the [MIT License](LICENSE).

<br>

<div align="center">

### Know which identity you're using **before** you code

<sub>Made for developers juggling too many identities.</sub>

</div>
