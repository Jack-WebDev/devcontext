package application

import (
	"sort"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
)

func (s *Service) getTrustCenter() (TrustCenterState, error) {
	contexts, err := s.dependencies.Contexts.List()
	if err != nil {
		return TrustCenterState{}, err
	}
	bindings, err := s.dependencies.Projects.List()
	if err != nil {
		return TrustCenterState{}, err
	}

	contextNames := make(map[devcontext.ID]string, len(contexts))
	protections := make([]TrustContextProtection, 0, len(contexts))
	boundariesByTool := make(map[codingtool.ID]TrustIntegrationBoundary)
	for _, ctx := range contexts {
		contextNames[ctx.ID] = ctx.Name
		state := s.contextState(ctx)
		protections = append(protections, trustContextProtection(state))
		toolID := codingtool.ID(state.Tool.ID)
		if _, seen := boundariesByTool[toolID]; !seen {
			boundariesByTool[toolID] = s.trustIntegrationBoundary(toolID, state.Tool.Name)
		}
	}

	mappings := make([]TrustProjectMapping, 0, len(bindings))
	for _, binding := range bindings {
		contextName, ok := contextNames[binding.ContextID]
		if !ok {
			contextName = "Unavailable context"
		}
		mappings = append(mappings, TrustProjectMapping{
			Project: projectState(binding.ProjectPath), ContextID: binding.ContextID.String(), ContextName: contextName,
		})
	}

	boundaries := make([]TrustIntegrationBoundary, 0, len(boundariesByTool))
	for _, boundary := range boundariesByTool {
		boundaries = append(boundaries, boundary)
	}
	sort.Slice(boundaries, func(i, j int) bool { return boundaries[i].ToolID < boundaries[j].ToolID })

	return TrustCenterState{
		Contexts:        protections,
		ProjectMappings: mappings,
		CredentialSync: TrustCredentialSyncProtection{
			Enabled: false,
			Message: "Dev Context does not sync credentials. Provider credentials remain in local context-owned storage.",
		},
		IntegrationBoundaries: boundaries,
	}, nil
}

func trustContextProtection(context ContextState) TrustContextProtection {
	providers := make([]TrustProviderProtection, 0)
	for _, provider := range context.Providers {
		if !provider.Enabled {
			continue
		}
		providers = append(providers, TrustProviderProtection{
			ID: provider.ID, Name: provider.Name,
			Isolation: trustIsolationProtection(context.Confidence.Checks, func(check LaunchConfidenceCheck) bool {
				return check.Component == LaunchConfidenceCheckIsolation && check.ProviderID == provider.ID
			}),
		})
	}
	return TrustContextProtection{
		ID: context.ID, Name: context.Name, Providers: providers,
		Tool: TrustCodingToolProtection{
			ID: context.Tool.ID, Name: context.Tool.Name,
			Isolation: trustIsolationProtection(context.Confidence.Checks, func(check LaunchConfidenceCheck) bool {
				return check.Component == LaunchConfidenceCheckIsolation && check.ToolID == context.Tool.ID
			}),
		},
	}
}

func trustIsolationProtection(checks []LaunchConfidenceCheck, matches func(LaunchConfidenceCheck) bool) TrustIsolationProtection {
	for _, check := range checks {
		if matches(check) {
			return TrustIsolationProtection{Status: check.Severity, Message: check.Message}
		}
	}
	return TrustIsolationProtection{Status: LaunchConfidenceBlocked, Message: "Isolation readiness could not be determined."}
}

func (s *Service) trustIntegrationBoundary(toolID codingtool.ID, toolName string) TrustIntegrationBoundary {
	registered, ok := s.dependencies.ToolRegistry.Lookup(toolID)
	if !ok {
		return TrustIntegrationBoundary{ToolID: string(toolID), ToolName: toolName, Message: "This coding tool is not registered, so Dev Context does not export integration data."}
	}
	_, consumesStatusData := registered.Integration.(codingtool.StatusDataConsumer)
	if !consumesStatusData {
		return TrustIntegrationBoundary{ToolID: string(toolID), ToolName: registered.DisplayName, Message: "Dev Context does not export integration data to this coding tool."}
	}
	return TrustIntegrationBoundary{
		ToolID: string(toolID), ToolName: registered.DisplayName, StatusDataAvailable: true,
		Message: "Dev Context may write safe project, context, provider-identity, and isolation-status data into this coding tool's isolated storage. Credentials, commands, and environment values are excluded.",
	}
}
