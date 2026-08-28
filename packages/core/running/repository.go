package running

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

// Repository persists running environment records in one local TOML file.
type Repository struct {
	path string
}

// RefreshResult reports persisted environments and records that changed to stopped.
type RefreshResult struct {
	Environments []Environment
	Stopped      []Environment
}

// NewRepository creates a repository backed by one TOML file.
func NewRepository(path string) Repository {
	return Repository{path: path}
}

// IsZero reports whether no persistence location has been configured.
func (r Repository) IsZero() bool {
	return r.path == ""
}

// List returns all stored running environments in deterministic identity order.
func (r Repository) List() ([]Environment, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Environment{}, nil
		}
		return nil, fmt.Errorf("read running environments %q: %w", r.path, err)
	}

	var document document
	if _, err := toml.Decode(string(data), &document); err != nil {
		return nil, fmt.Errorf("decode running environments %q: %w", r.path, err)
	}
	environments := make([]Environment, 0, len(document.Environments))
	for index, entry := range document.Environments {
		environment, err := environmentFromTOML(entry)
		if err != nil {
			return nil, fmt.Errorf("decode running environments %q: environments[%d]: %w", r.path, index, err)
		}
		environments = append(environments, environment)
	}
	sortEnvironments(environments)
	return environments, nil
}

// Record creates or updates the record for one project/context launch identity.
// A repeated launch of the same project and context updates its observed state
// while preserving the environment ID; a different context creates a separate record.
func (r Repository) Record(environment Environment) (Environment, error) {
	if r.IsZero() {
		return Environment{}, fmt.Errorf("record running environment: repository is not configured")
	}
	if err := validateEnvironment(environment); err != nil {
		return Environment{}, fmt.Errorf("record running environment: %w", err)
	}

	environments, err := r.List()
	if err != nil {
		return Environment{}, err
	}
	for index, existing := range environments {
		if existing.Project.Path == environment.Project.Path && existing.Context.ID == environment.Context.ID {
			environment.ID = existing.ID
			environments[index] = environment
			if err := r.write(environments); err != nil {
				return Environment{}, err
			}
			return environment, nil
		}
	}

	id, err := newID()
	if err != nil {
		return Environment{}, fmt.Errorf("create running environment ID: %w", err)
	}
	environment.ID = id
	environments = append(environments, environment)
	if err := r.write(environments); err != nil {
		return Environment{}, err
	}
	return environment, nil
}

// RefreshProcessStates marks running environments stopped when their recorded
// PID is no longer active. Records without a PID remain unchanged because their
// process state cannot be observed safely.
func (r Repository) RefreshProcessStates(inspector ProcessInspector) (RefreshResult, error) {
	if r.IsZero() {
		return RefreshResult{}, fmt.Errorf("refresh running environments: repository is not configured")
	}
	if inspector == nil {
		return RefreshResult{}, fmt.Errorf("refresh running environments: process inspector is required")
	}
	environments, err := r.List()
	if err != nil {
		return RefreshResult{}, err
	}
	stopped := make([]Environment, 0)
	changed := false
	for index := range environments {
		environment := &environments[index]
		if environment.Process.State != ProcessStateRunning || environment.Process.PID == nil {
			continue
		}
		running, err := inspector.IsRunning(*environment.Process.PID)
		if err != nil {
			return RefreshResult{}, err
		}
		if running {
			continue
		}
		environment.Process.State = ProcessStateStopped
		if environment.Session.State == SessionStateActive {
			environment.Session.State = SessionStateEnded
		}
		stopped = append(stopped, *environment)
		changed = true
	}
	if changed {
		if err := r.write(environments); err != nil {
			return RefreshResult{}, err
		}
	}
	return RefreshResult{Environments: environments, Stopped: stopped}, nil
}

type document struct {
	Environments []environmentTOML `toml:"environments"`
}

type environmentTOML struct {
	ID               string `toml:"id"`
	ProjectPath      string `toml:"project_path"`
	ProjectName      string `toml:"project_name"`
	ContextID        string `toml:"context_id"`
	ContextName      string `toml:"context_name"`
	ToolID           string `toml:"tool_id"`
	ToolName         string `toml:"tool_name"`
	StartedAt        string `toml:"started_at"`
	ProcessState     string `toml:"process_state"`
	ProcessPID       *int   `toml:"process_pid,omitempty"`
	SessionID        string `toml:"session_id,omitempty"`
	SessionState     string `toml:"session_state"`
	LaunchSource     string `toml:"launch_source"`
	ResolutionSource string `toml:"resolution_source"`
}

func environmentFromTOML(entry environmentTOML) (Environment, error) {
	if entry.ID == "" {
		return Environment{}, fmt.Errorf("missing environment ID")
	}
	contextID, err := devcontext.NewID(entry.ContextID)
	if err != nil {
		return Environment{}, fmt.Errorf("invalid context ID: %w", err)
	}
	startedAt, err := parseStartedAt(entry.StartedAt)
	if err != nil {
		return Environment{}, err
	}
	environment := Environment{
		ID:        ID(entry.ID),
		Project:   ProjectIdentity{Path: project.Path(entry.ProjectPath), Name: entry.ProjectName},
		Context:   ContextIdentity{ID: contextID, Name: entry.ContextName},
		Tool:      ToolIdentity{ID: codingtool.ID(entry.ToolID), Name: entry.ToolName},
		StartedAt: startedAt,
		Process:   Process{PID: entry.ProcessPID, State: ProcessState(entry.ProcessState)},
		Session:   Session{ID: entry.SessionID, State: SessionState(entry.SessionState)},
		Launch:    LaunchIdentity{Source: launcher.InvocationSource(entry.LaunchSource), ResolutionSource: launcher.ResolutionSource(entry.ResolutionSource)},
	}
	if err := validateEnvironment(environment); err != nil {
		return Environment{}, err
	}
	return environment, nil
}

func (r Repository) write(environments []Environment) error {
	sortEnvironments(environments)
	document := document{Environments: make([]environmentTOML, len(environments))}
	for index, environment := range environments {
		document.Environments[index] = environmentTOML{
			ID: string(environment.ID), ProjectPath: string(environment.Project.Path), ProjectName: environment.Project.Name,
			ContextID: environment.Context.ID.String(), ContextName: environment.Context.Name,
			ToolID: string(environment.Tool.ID), ToolName: environment.Tool.Name,
			StartedAt:    environment.StartedAt.UTC().Format(timeFormat),
			ProcessState: string(environment.Process.State), ProcessPID: environment.Process.PID,
			SessionID: environment.Session.ID, SessionState: string(environment.Session.State),
			LaunchSource: string(environment.Launch.Source), ResolutionSource: string(environment.Launch.ResolutionSource),
		}
	}

	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(document); err != nil {
		return fmt.Errorf("encode running environments %q: %w", r.path, err)
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o700); err != nil {
		return fmt.Errorf("create running environments directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(r.path), ".running-*.toml")
	if err != nil {
		return fmt.Errorf("create running environments temporary file: %w", err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)
	if _, err := file.Write(encoded.Bytes()); err != nil {
		file.Close()
		return fmt.Errorf("write running environments temporary file: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("protect running environments temporary file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close running environments temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, r.path); err != nil {
		return fmt.Errorf("replace running environments %q: %w", r.path, err)
	}
	return nil
}

const timeFormat = "2006-01-02T15:04:05.999999999Z07:00"

func parseStartedAt(value string) (time.Time, error) {
	startedAt, err := time.Parse(timeFormat, value)
	if err != nil || startedAt.IsZero() {
		return time.Time{}, fmt.Errorf("invalid started time")
	}
	return startedAt.UTC(), nil
}

func validateEnvironment(environment Environment) error {
	if environment.Project.Path == "" || environment.Project.Name == "" || environment.Context.ID.String() == "" || environment.Context.Name == "" || environment.Tool.ID == "" || environment.Tool.Name == "" || environment.StartedAt.IsZero() {
		return fmt.Errorf("project, context, tool, and started time are required")
	}
	if !validProcessState(environment.Process.State) || !validSessionState(environment.Session.State) {
		return fmt.Errorf("invalid process or session state")
	}
	if environment.Launch.Source == "" || environment.Launch.ResolutionSource == "" {
		return fmt.Errorf("launch source and resolution source are required")
	}
	return nil
}

func validProcessState(state ProcessState) bool {
	return state == ProcessStateUnknown || state == ProcessStateRunning || state == ProcessStateStopped
}

func validSessionState(state SessionState) bool {
	return state == SessionStateUnknown || state == SessionStateActive || state == SessionStateEnded
}

func sortEnvironments(environments []Environment) {
	sort.Slice(environments, func(i, j int) bool {
		if environments[i].Project.Path == environments[j].Project.Path {
			return environments[i].Context.ID.String() < environments[j].Context.ID.String()
		}
		return environments[i].Project.Path < environments[j].Project.Path
	})
}

func newID() (ID, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return ID(hex.EncodeToString(bytes)), nil
}
