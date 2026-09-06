import { useEffect, useState } from "react";

import type {
	ApiResult,
	ContextDetailsState,
	ContextMetadataExport,
	DisplayError,
	DuplicateContextRequest,
	DuplicateContextResult,
	ExportContextMetadataRequest,
	ImportContextMetadataRequest,
	ImportContextMetadataResult,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import {
	Sheet,
	SheetContent,
	SheetDescription,
	SheetHeader,
	SheetTitle,
} from "../ui/sheet.js";
import { ContextField } from "./ContextField";
import {
	contextMetadataImportReview,
	parseContextMetadataExport,
} from "./context-transfer";

interface ContextDetailsDrawerProps {
	contextId: string;
	onClose: () => void;
	load: (id: string) => Promise<ApiResult<ContextDetailsState>>;
	duplicate: (
		request: DuplicateContextRequest,
	) => Promise<ApiResult<DuplicateContextResult>>;
	exportMetadata: (
		request: ExportContextMetadataRequest,
	) => Promise<ApiResult<ContextMetadataExport>>;
	importMetadata: (
		request: ImportContextMetadataRequest,
	) => Promise<ApiResult<ImportContextMetadataResult>>;
	chooseImportFile: () => Promise<ApiResult<string | undefined>>;
}

export function ContextDetailsDrawer({
	contextId,
	onClose,
	load,
	duplicate,
	exportMetadata,
	importMetadata,
	chooseImportFile,
}: ContextDetailsDrawerProps) {
	const [result, setResult] = useState<ApiResult<ContextDetailsState>>();
	const [duplicateID, setDuplicateID] = useState("");
	const [duplicateName, setDuplicateName] = useState("");
	const [duplicateError, setDuplicateError] = useState<DisplayError>();
	const [duplicatePending, setDuplicatePending] = useState(false);
	const [exportedMetadata, setExportedMetadata] = useState("");
	const [exportError, setExportError] = useState<DisplayError>();
	const [exportPending, setExportPending] = useState(false);
	const [importedMetadata, setImportedMetadata] =
		useState<ContextMetadataExport>();
	const [importError, setImportError] = useState<string>();
	const [importPending, setImportPending] = useState(false);

	useEffect(() => {
		void load(contextId).then(setResult);
	}, [contextId, load]);

	useEffect(() => {
		if (result?.ok) {
			setDuplicateID(`${result.data.context.id}-copy`);
			setDuplicateName(`${result.data.context.name} copy`);
		}
	}, [result]);

	async function submitDuplicate() {
		if (!result?.ok) return;

		setDuplicatePending(true);
		setDuplicateError(undefined);
		const duplicated = await duplicate({
			sourceContextId: result.data.context.id,
			contextId: duplicateID,
			name: duplicateName,
		});
		setDuplicatePending(false);

		if (!duplicated.ok) {
			setDuplicateError(duplicated.error);
			return;
		}
		onClose();
	}

	async function prepareExport() {
		if (!result?.ok) return;
		setExportPending(true);
		setExportError(undefined);
		const exported = await exportMetadata({
			contextId: result.data.context.id,
		});
		setExportPending(false);
		if (!exported.ok) {
			setExportError(exported.error);
			return;
		}
		setExportedMetadata(JSON.stringify(exported.data, null, 2));
	}

	async function chooseImport() {
		setImportError(undefined);
		const selected = await chooseImportFile();
		if (!selected.ok) {
			setImportError(selected.error.message);
			return;
		}
		if (!selected.data) return;
		try {
			setImportedMetadata(parseContextMetadataExport(selected.data));
		} catch {
			setImportedMetadata(undefined);
			setImportError("Choose a valid context metadata export file.");
		}
	}

	async function submitImport() {
		if (!importedMetadata) return;
		setImportPending(true);
		setImportError(undefined);
		const imported = await importMetadata({
			export: importedMetadata,
		});
		setImportPending(false);
		if (!imported.ok) {
			setImportError(imported.error.message);
			return;
		}
		onClose();
	}

	return (
		<Sheet open onOpenChange={(open) => !open && onClose()}>
			<SheetContent>
				<SheetHeader>
					<SheetTitle>Context details</SheetTitle>
					<SheetDescription>
						Backend-owned context information.
					</SheetDescription>
				</SheetHeader>
				<div className="space-y-3 px-8 pb-8">
					{!result ? (
						<p>Loading context details...</p>
					) : !result.ok ? (
						<p className="text-destructive">{result.error.message}</p>
					) : (
						<>
							<Detail label="Name" value={result.data.context.name} />
							<Detail label="Location" value={result.data.location} />
							<Detail
								label="Created"
								value={new Date(result.data.createdAt).toLocaleString()}
							/>
							<Detail
								label="Projects"
								value={String(result.data.projectCount)}
							/>
							<Detail
								label="Coding tool"
								value={result.data.context.tool.name}
							/>
							<Detail
								label="Providers"
								value={
									result.data.enabledProviders
										.map((provider) => provider.name)
										.join(", ") || "None"
								}
							/>
							<section
								className="space-y-3 border-t border-border pt-4"
								aria-labelledby="duplicate-context-heading"
							>
								<div>
									<h3 id="duplicate-context-heading" className="font-medium">
										Duplicate context
									</h3>
									<p className="mt-1 text-sm text-muted-foreground">
										Copies context metadata, provider settings, and coding-tool
										settings into a new isolated context. Credentials are not
										copied.
									</p>
								</div>
								<ContextField
									label="New name"
									value={duplicateName}
									onChange={setDuplicateName}
								/>
								<ContextField
									label="New ID"
									value={duplicateID}
									onChange={setDuplicateID}
								/>
								{duplicateError ? (
									<p className="text-destructive">{duplicateError.message}</p>
								) : null}
								<Button
									type="button"
									disabled={duplicatePending || !duplicateID || !duplicateName}
									onClick={() => void submitDuplicate()}
								>
									{duplicatePending ? "Duplicating..." : "Duplicate context"}
								</Button>
							</section>
							<section
								className="space-y-3 border-t border-border pt-4"
								aria-labelledby="context-transfer-heading"
							>
								<div>
									<h3 id="context-transfer-heading" className="font-medium">
										Import and export
									</h3>
									<p className="mt-1 text-sm text-muted-foreground">
										Exports include context metadata and non-secret provider and
										coding-tool settings. Credentials are never included or
										imported.
									</p>
								</div>
								<Button
									type="button"
									variant="outline"
									disabled={exportPending}
									onClick={() => void prepareExport()}
								>
									{exportPending
										? "Preparing export..."
										: "Prepare safe export"}
								</Button>
								{exportError ? (
									<p className="text-destructive">{exportError.message}</p>
								) : null}
								{exportedMetadata ? (
									<label className="block text-sm">
										Safe context metadata
										<textarea
											aria-label="Safe context metadata export"
											className="mt-1 min-h-40 w-full border p-2 font-mono text-xs"
											value={exportedMetadata}
											readOnly
										/>
									</label>
								) : null}
								<Button
									type="button"
									variant="outline"
									disabled={importPending}
									onClick={() => void chooseImport()}
								>
									Choose context export file
								</Button>
								{importedMetadata ? (
									<ContextMetadataImportReviewCard
										exported={importedMetadata}
										availableTools={result.data.context.availableTools.map((tool) => tool.id)}
										availableProviders={result.data.context.providers.map(
											(provider) => provider.id,
										)}
									/>
								) : null}
								{importError ? (
									<p className="text-destructive">{importError}</p>
								) : null}
								<Button
									type="button"
									disabled={importPending || !importedMetadata}
									onClick={() => void submitImport()}
								>
									{importPending ? "Importing..." : "Confirm import"}
								</Button>
							</section>
						</>
					)}
				</div>
			</SheetContent>
		</Sheet>
	);
}

function ContextMetadataImportReviewCard({
	exported,
	availableTools,
	availableProviders,
}: {
	exported: ContextMetadataExport;
	availableTools: string[];
	availableProviders: string[];
}) {
	const review = contextMetadataImportReview(exported, {
		toolIds: availableTools,
		providerIds: availableProviders,
	});
	return (
		<section
			className="space-y-2 rounded-md border border-border p-3 text-sm"
			aria-label="Import review"
		>
			<h4 className="font-medium">Review import</h4>
			<Detail label="Name" value={review.name} />
			<Detail label="Tools" value={review.tools.join(", ") || "None"} />
			<Detail
				label="Preferences"
				value={`Default coding tool: ${review.defaultTool || "None"}`}
			/>
			<Detail
				label="Project paths"
				value="Not included. This context will not be linked to any projects."
			/>
			<Detail
				label="Missing integrations"
				value={
					review.missingIntegrations.join(", ") ||
					"None detected on this device."
				}
			/>
			<p className="text-muted-foreground">
				Confirming creates a new isolated context with a generated internal ID.
			</p>
		</section>
	);
}

function Detail({ label, value }: { label: string; value: string }) {
	return (
		<div>
			<p className="text-xs uppercase text-muted-foreground">{label}</p>
			<p className="mt-1 break-all">{value}</p>
		</div>
	);
}
