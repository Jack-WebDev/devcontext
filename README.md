# Dev Context

Dev Context is a local desktop launcher for keeping development identities separate on the same machine.

The MVP focuses on launching VS Code with isolated Personal and Company environments for Claude Code, Codex, and editor state. The longer-term goal is to make `devctx` the entry point for a complete development context, including accounts, environment variables, editor state, and other developer tooling.

## Status

This project is early-stage. The current codebase is a Wails desktop app scaffold with the product plan kept outside the public repository.

## Tech Stack

- Go
- Wails
- React
- Vite
- TypeScript
- Tailwind CSS

## Development

Install frontend dependencies:

```bash
cd frontend
npm install
```

Run the desktop app in development mode from the repository root:

```bash
wails dev
```

## Building

Build a production desktop package:

```bash
wails build
```

## Contributing

Issues and pull requests are welcome. Keep changes focused, include a clear description of the behavior being changed, and run the relevant build or test command before opening a pull request.

## License

MIT
