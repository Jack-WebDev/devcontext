package version

const developmentVersion = "dev"

// These variables are intentionally package variables so release builds can
// inject values with Go ldflags.
var (
	Version = developmentVersion
	Commit  = "unknown"
	Date    = "unknown"
)

// Info is the build identity rendered by user-facing adapters.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Current returns normalized build metadata.
func Current() Info {
	info := Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
	if info.Version == "" {
		info.Version = developmentVersion
	}
	if info.Commit == "" {
		info.Commit = "unknown"
	}
	if info.Date == "" {
		info.Date = "unknown"
	}
	return info
}

// Render returns the stable CLI version response.
func Render(info Info) string {
	if info.Version == "" {
		info.Version = developmentVersion
	}
	return "devctx " + info.Version + "\n"
}
