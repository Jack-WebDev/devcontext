package project

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	devcontext "devctx/packages/core/context"
)

// RecentProject records a successful launch without changing the project's
// remembered context binding.
type RecentProject struct {
	ProjectPath    Path
	ContextID      devcontext.ID
	LastLaunchedAt time.Time
}

// RecentRepository persists recent successful project launches.
type RecentRepository struct {
	path string
}

// NewRecentRepository creates a repository backed by one TOML file.
func NewRecentRepository(path string) RecentRepository {
	return RecentRepository{path: path}
}

// IsZero reports whether no persistence location has been configured.
func (r RecentRepository) IsZero() bool {
	return r.path == ""
}

// List returns all stored recent launches.
func (r RecentRepository) List() ([]RecentProject, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read recent projects %q: %w", r.path, err)
	}

	var document recentProjectsDocument
	if _, err := toml.Decode(string(data), &document); err != nil {
		return nil, fmt.Errorf("decode recent projects %q: %w", r.path, err)
	}

	projects := make([]RecentProject, 0, len(document.Projects))
	for index, entry := range document.Projects {
		if entry.Path == "" || entry.Context == "" || entry.LastLaunchedAt.IsZero() {
			return nil, fmt.Errorf("decode recent projects %q: invalid projects[%d]", r.path, index)
		}
		contextID, err := devcontext.NewID(entry.Context)
		if err != nil {
			return nil, fmt.Errorf("decode recent projects %q: invalid projects[%d] context: %w", r.path, index, err)
		}
		projects = append(projects, RecentProject{ProjectPath: Path(entry.Path), ContextID: contextID, LastLaunchedAt: entry.LastLaunchedAt.UTC()})
	}
	return projects, nil
}

// Record stores the latest successful launch for a project.
func (r RecentRepository) Record(projectPath Path, contextID devcontext.ID, launchedAt time.Time) error {
	if projectPath == "" || contextID.String() == "" || launchedAt.IsZero() {
		return fmt.Errorf("record recent project: invalid project, context, or launch time")
	}

	projects, err := r.List()
	if err != nil {
		return err
	}
	recent := RecentProject{ProjectPath: projectPath, ContextID: contextID, LastLaunchedAt: launchedAt.UTC()}
	replaced := false
	for index := range projects {
		if projects[index].ProjectPath == projectPath {
			projects[index] = recent
			replaced = true
			break
		}
	}
	if !replaced {
		projects = append(projects, recent)
	}
	return WriteRecentProjectsFile(r.path, projects)
}

type recentProjectsDocument struct {
	Projects []recentProjectTOML `toml:"projects"`
}

type recentProjectTOML struct {
	Path           string    `toml:"path"`
	Context        string    `toml:"context"`
	LastLaunchedAt time.Time `toml:"last_launched_at"`
}

// WriteRecentProjectsFile writes recent projects atomically in deterministic
// project-path order.
func WriteRecentProjectsFile(path string, projects []RecentProject) error {
	ordered := append([]RecentProject(nil), projects...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ProjectPath < ordered[j].ProjectPath })

	var builder strings.Builder
	for index, recent := range ordered {
		if index > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString("[[projects]]\n")
		builder.WriteString("path = ")
		builder.WriteString(strconv.Quote(string(recent.ProjectPath)))
		builder.WriteString("\ncontext = ")
		builder.WriteString(strconv.Quote(recent.ContextID.String()))
		builder.WriteString("\nlast_launched_at = ")
		builder.WriteString(recent.LastLaunchedAt.UTC().Format(time.RFC3339Nano))
		builder.WriteString("\n")
	}

	return writeProjectBindingsFileAtomically(path, func(file *os.File) error {
		_, err := file.WriteString(builder.String())
		return err
	})
}
