import type { ContextMetadataExport } from "../../lib/devctx-api";

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

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}
