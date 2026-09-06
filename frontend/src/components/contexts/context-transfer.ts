import type { ContextMetadataExport } from "../../lib/devctx-api";

export interface ContextMetadataImportReview {
	name: string;
	tools: string[];
	defaultTool: string;
	enabledProviders: string[];
	missingIntegrations: string[];
}

// parseContextMetadataExport accepts the outer shape needed to call the
// import API. The backend remains authoritative for version and integration
// validation.
export function parseContextMetadataExport(
	value: string,
): ContextMetadataExport {
	const parsed: unknown = JSON.parse(value);
	if (
		!isRecord(parsed) ||
		!Number.isInteger(parsed.version) ||
		!isRecord(parsed.context)
	) {
		throw new Error("invalid context metadata export");
	}
	return parsed as unknown as ContextMetadataExport;
}

// contextMetadataImportReview keeps the pre-commit review presentation-only.
// The receiving application remains authoritative when it validates the file.
export function contextMetadataImportReview(
	exported: ContextMetadataExport,
	available: { toolIds: string[]; providerIds: string[] },
): ContextMetadataImportReview {
	const tools = unique([
		exported.context.launchTarget.defaultTool,
		...exported.context.launchTarget.tools.map((tool) => tool.id),
	]);
	const enabledProviders = exported.context.providers
		.filter((provider) => provider.enabled)
		.map((provider) => provider.id);
	const providerIDs = exported.context.providers.map((provider) => provider.id);
	const availableTools = new Set(available.toolIds);
	const availableProviders = new Set(available.providerIds);
	return {
		name: exported.context.name,
		tools,
		defaultTool: exported.context.launchTarget.defaultTool,
		enabledProviders,
		missingIntegrations: unique([
			...tools.filter((id) => !availableTools.has(id)),
			...providerIDs.filter((id) => !availableProviders.has(id)),
		]),
	};
}

function unique(values: string[]) {
	return [...new Set(values.filter(Boolean))];
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}
