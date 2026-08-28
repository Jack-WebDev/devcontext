package application

import (
	"encoding/json"
	"fmt"
	"os"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/launcher"
)

const codingToolStatusSchemaVersion = 1

// CodingToolStatusData is the safe, local status document that a registered
// coding-tool integration may consume. It intentionally excludes credentials,
// commands, environment values, and context storage paths.
type CodingToolStatusData struct {
	SchemaVersion int                            `json:"schemaVersion"`
	Project       ProjectState                   `json:"project"`
	Context       CodingToolStatusContext        `json:"context"`
	Providers     []CodingToolStatusProvider     `json:"providers"`
	Isolation     CodingToolStatusIsolationState `json:"isolation"`
}

// CodingToolStatusContext identifies the immutable context and selected tool.
type CodingToolStatusContext struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Tool ToolOption `json:"tool"`
}

// CodingToolStatusProvider exposes only an enabled provider's safe identity
// metadata. Provider readiness, credentials, and integration configuration
// remain private to Dev Context.
type CodingToolStatusProvider struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Identity ProviderIdentityState `json:"identity"`
}

// CodingToolStatusIsolationState summarizes backend-derived isolation
// readiness without disclosing filesystem locations.
type CodingToolStatusIsolationState struct {
	Status  LaunchConfidenceStatus `json:"status"`
	Message string                 `json:"message"`
}

func (s *Service) exportCodingToolStatus(plan launcher.LaunchPlan) error {
	registered, ok := s.dependencies.ToolRegistry.Lookup(plan.Tool.ID)
	if !ok {
		return fmt.Errorf("selected coding tool %q is not registered", plan.Tool.ID)
	}
	consumer, ok := registered.Integration.(codingtool.StatusDataConsumer)
	if !ok {
		return nil
	}

	statusPath, err := codingtool.StatusDataPath(codingtool.ContextPaths{
		RootDir:    plan.ContextPaths.RootDir,
		StorageDir: plan.ContextPaths.ToolStorageDir(plan.Tool.ID),
	}, consumer)
	if err != nil {
		return fmt.Errorf("resolve %s status data path: %w", plan.Tool.DisplayName, err)
	}

	data, err := json.Marshal(codingToolStatusData(plan, s.contextState(plan.Context)))
	if err != nil {
		return fmt.Errorf("encode %s status data: %w", plan.Tool.DisplayName, err)
	}
	if err := os.WriteFile(statusPath, data, s.dependencies.StoragePermissions.FileMode()); err != nil {
		return fmt.Errorf("write %s status data: %w", plan.Tool.DisplayName, err)
	}
	if err := s.dependencies.StoragePermissions.ApplyFile(statusPath); err != nil {
		return fmt.Errorf("secure %s status data: %w", plan.Tool.DisplayName, err)
	}
	return nil
}

func codingToolStatusData(plan launcher.LaunchPlan, context ContextState) CodingToolStatusData {
	providers := make([]CodingToolStatusProvider, 0)
	for _, provider := range context.Providers {
		if provider.Enabled {
			providers = append(providers, CodingToolStatusProvider{
				ID:       provider.ID,
				Name:     provider.Name,
				Identity: provider.Identity,
			})
		}
	}
	return CodingToolStatusData{
		SchemaVersion: codingToolStatusSchemaVersion,
		Project:       projectState(plan.ProjectPath),
		Context: CodingToolStatusContext{
			ID:   context.ID,
			Name: context.Name,
			Tool: ToolOption{ID: string(plan.Tool.ID), Name: plan.Tool.DisplayName},
		},
		Providers: providers,
		Isolation: codingToolStatusIsolation(context.Confidence),
	}
}

func codingToolStatusIsolation(confidence LaunchConfidenceState) CodingToolStatusIsolationState {
	result := CodingToolStatusIsolationState{
		Status:  LaunchConfidenceReady,
		Message: "Context isolation is ready.",
	}
	for _, check := range confidence.Checks {
		if check.Component != LaunchConfidenceCheckIsolation || launchConfidenceSeverityRank(check.Severity) <= launchConfidenceSeverityRank(result.Status) {
			continue
		}
		result.Status = check.Severity
		result.Message = check.Message
	}
	return result
}

func launchConfidenceSeverityRank(status LaunchConfidenceStatus) int {
	switch status {
	case LaunchConfidenceBlocked:
		return 2
	case LaunchConfidenceNeedsAttention:
		return 1
	default:
		return 0
	}
}
