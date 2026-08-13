# Dev Context

Dev Context is a local desktop launcher for keeping development identities separate on the same machine.

It is built for developers who move between personal projects, company repositories, freelance work, and client code on one workstation. Instead of relying on whatever account or editor state happens to be active globally, Dev Context makes the active development context explicit before launching your tools.

## Goals

- Prevent accidental use of the wrong AI account or subscription.
- Keep personal and company editor state isolated.
- Support multiple active development contexts on the same machine.
- Provide a small local-first app with no hosted service requirement.

## Tech Stack

- Go
- Wails
- React
- Vite
- TypeScript
- Tailwind CSS
- Shadcn UI

## Requirements

- Go 1.25 or newer
- Node.js 22 or newer
- npm
- Wails CLI v2

## Development

Install frontend dependencies:

```bash
cd frontend
npm install
```

Run the desktop app from the repository root:

```bash
wails dev
```

Run the core checks used by CI:

```bash
cd frontend
npm ci
npm run build
cd ..
go test ./...
```

## Building

Build a production desktop package:

```bash
wails build
```

Build a release artifact with CLI version metadata:

```bash
scripts/build-release.sh 1.0.0
```

The built `devctx` executable also acts as the desktop entry point. `devctx --version` prints the installed version without initializing local storage.

## Release Checks

Run startup benchmarks:

```bash
scripts/benchmark-startup.sh
```

Release evidence templates live in [`docs/release`](docs/release).

## Contributing

Issues and pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening larger changes.

For security reports, see [SECURITY.md](SECURITY.md).

## License

Dev Context is released under the [MIT License](LICENSE).
