package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"devctx/packages/application"
	"devctx/packages/core/cli"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
	"devctx/packages/core/version"
	"devctx/packages/wailsapp"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	args := os.Args[1:]
	if shouldRunCLI(args) {
		os.Exit(int(runCLI(args)))
	}

	service, err := application.NewService()
	if err != nil {
		fmt.Fprint(os.Stderr, application.NewError(err).Error()+"\n")
		os.Exit(1)
	}
	app := wailsapp.New(service)

	err = wails.Run(&options.App{
		Title:  "devctx",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func shouldRunCLI(args []string) bool {
	if len(args) == 0 {
		return false
	}
	for _, arg := range args {
		if arg == "--debug" {
			return true
		}
	}
	if args[0] == string(cli.CommandContext) || args[0] == string(cli.CommandProject) {
		return true
	}
	if args[0] == "--version" {
		return true
	}
	for _, arg := range args {
		if arg == "--context" || arg == "--personal" || arg == "--company" {
			return true
		}
	}
	return false
}

func runCLI(args []string) cli.ExitCode {
	parsedArgs, debug := cli.SplitDebugFlag(args)
	if _, err := cli.Parse(parsedArgs); err != nil {
		fmt.Fprint(os.Stderr, cli.RenderError(err, debug))
		return cli.ExitCodeForError(err)
	}
	if len(parsedArgs) == 1 && parsedArgs[0] == "--version" {
		fmt.Fprint(os.Stdout, version.Render(version.Current()))
		return cli.ExitSuccess
	}

	paths := filesystem.NewDefaultPlatformPaths()
	layout, err := config.InitializeDevContextHome(paths)
	if err != nil {
		fmt.Fprint(os.Stderr, cli.RenderError(err, debug))
		return cli.ExitCodeForError(err)
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprint(os.Stderr, cli.RenderError(err, debug))
		return cli.ExitInternalError
	}

	runner := cli.Runner{
		Contexts:          devcontext.NewRepository(layout.ContextsDir),
		Projects:          project.NewRepository(filepath.Join(layout.HomeDir, "projects.toml"), paths),
		WorkingDirectory:  workingDirectory,
		Paths:             paths,
		ProviderRegistry:  provider.DefaultRegistry(),
		Editor:            editor.VSCodeEditor{},
		ProcessLauncher:   launcher.NativeProcessLauncher{},
		ParentEnvironment: os.Environ(),
		DetachMode:        launcher.DetachModeDetached,
		Debug:             debug,
		Logger:            devlog.NewLocalLogger(layout.LogsDir, filesystem.NewDefaultStoragePermissions(), nil),
	}
	result := runner.Run(parsedArgs)
	if err := result.Write(os.Stdout, os.Stderr); err != nil {
		fmt.Fprint(os.Stderr, cli.RenderError(err, debug))
		return cli.ExitInternalError
	}
	return result.Code
}
