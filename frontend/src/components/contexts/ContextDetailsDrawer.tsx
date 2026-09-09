import { useEffect, useState } from "react";

import type {
	ApiResult,
	ContextDetailsState,
	ContextMetadataExport,
	ContextMetadataExportOptions,
	DisplayError,
	DuplicateContextRequest,
	DuplicateContextResult,
	ExportContextMetadataRequest,
	ImportContextMetadataRequest,
	ImportContextMetadataResult,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Checkbox } from "../ui/checkbox.js";
import { Skeleton } from "../ui/skeleton.js";
import { Textarea } from "../ui/textarea.js";
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
	contextNames: string[];
}

export function ContextDetailsDrawer({
	contextId,
	onClose,
	load,
	duplicate,
	exportMetadata,
	importMetadata,
	chooseImportFile,
	contextNames,
}: ContextDetailsDrawerProps) {
	const [result, setResult] = useState<ApiResult<ContextDetailsState>>();
	const [duplicateID, setDuplicateID] = useState("");
	const [duplicateName, setDuplicateName] = useState("");
	const [duplicateError, setDuplicateError] = useState<DisplayError>();
	const [duplicatePending, setDuplicatePending] = useState(false);
	const [exportedMetadata, setExportedMetadata] = useState("");
	const [exportError, setExportError] = useState<DisplayError>();
	const [exportPending, setExportPending] = useState(false);
	const [exportOptions, setExportOptions] =
		useState<ContextMetadataExportOptions>({
			includeMetadata: true,
			includeProviderOptions: true,
			includeToolOptions: true,
		});
	const [importedMetadata, setImportedMetadata] =
		useState<ContextMetadataExport>();
	const [importName, setImportName] = useState("");
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
			options: exportOptions,
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
			const exported = parseContextMetadataExport(selected.data);
			setImportedMetadata(exported);
			setImportName(exported.context.name);
		} catch {
			setImportedMetadata(undefined);
			setImportName("");
			setImportError("Choose a valid context metadata export file.");
		}
	}

	async function submitImport() {
		if (!importedMetadata) return;
		setImportPending(true);
		setImportError(undefined);
		const imported = await importMetadata({
			name: importName.trim(),
			export: importedMetadata,
		});
		setImportPending(false);
		if (!imported.ok) {
			setImportError(imported.error.message);
			return;
		}
		onClose();
	}

	const nameConflict =
		importedMetadata !== undefined &&
		contextNames.some(
			(name) => normalizeContextName(name) === normalizeContextName(importName),
		);

	return (
		<Sheet open onOpenChange={(open) => !open && onClose()}>
			<SheetContent className="overflow-hidden">
				<SheetHeader className="border-b border-border/60 bg-muted/20">
					<SheetTitle>Context details</SheetTitle>
					<SheetDescription>
						Backend-owned context information.
					</SheetDescription>
				</SheetHeader>
				<div className="flex-1 space-y-5 overflow-y-auto px-8 pb-8">
					{!result ? (
						<div
							className="space-y-3"
							aria-label="Loading context details"
							aria-busy="true"
						>
							<Skeleton className="h-4 w-1/3" />
							<Skeleton className="h-4 w-2/3" />
							<Skeleton className="h-24 w-full" />
						</div>
					) : !result.ok ? (
						<p className="text-destructive">{result.error.message}</p>
					) : (
						<>
							<div className="grid gap-4 rounded-xl bg-muted/35 p-4 sm:grid-cols-2">
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
							</div>
							<section
								className="space-y-4 pt-2"
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
								<ExportOptions
									value={exportOptions}
									onChange={setExportOptions}
								/>
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
								className="space-y-4 border-t border-border/60 pt-6"
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
										<Textarea
											aria-label="Safe context metadata export"
											className="mt-2 min-h-40 font-mono text-xs"
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
									<>
										<ContextMetadataImportReviewCard
											exported={importedMetadata}
											name={importName}
											availableTools={result.data.context.availableTools.map(
												(tool) => tool.id,
											)}
											availableProviders={result.data.context.providers.map(
												(provider) => provider.id,
											)}
										/>
										{nameConflict ? (
											<ImportNameConflict
												name={importName}
												onNameChange={setImportName}
												onImportCopy={() =>
													setImportName(nextCopyName(importName, contextNames))
												}
												onCancel={() => {
													setImportedMetadata(undefined);
													setImportName("");
												}}
											/>
										) : null}
									</>
								) : null}
								{importError ? (
									<p className="text-destructive">{importError}</p>
								) : null}
								<Button
									type="button"
									disabled={
										importPending ||
										!importedMetadata ||
										!importName.trim() ||
										nameConflict
									}
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
	name,
	availableTools,
	availableProviders,
}: {
	exported: ContextMetadataExport;
	name: string;
	availableTools: string[];
	availableProviders: string[];
}) {
	const review = contextMetadataImportReview(exported, {
		toolIds: availableTools,
		providerIds: availableProviders,
	});
	return (
		<section
			className="space-y-2 rounded-xl bg-muted/35 p-4 text-sm"
			aria-label="Import review"
		>
			<h4 className="font-medium">Review import</h4>
			<Detail label="Name" value={name} />
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

function ExportOptions({
	value,
	onChange,
}: {
	value: ContextMetadataExportOptions;
	onChange: (value: ContextMetadataExportOptions) => void;
}) {
	const options: Array<{
		key: keyof ContextMetadataExportOptions;
		label: string;
	}> = [
		{ key: "includeMetadata", label: "Context details" },
		{ key: "includeProviderOptions", label: "Provider settings" },
		{ key: "includeToolOptions", label: "Coding-tool settings" },
	];
	return (
		<fieldset className="space-y-2 text-sm">
			<legend className="font-medium">Include in export</legend>
			{options.map((option) => (
				<label
					key={option.key}
					className="flex items-center gap-2.5 rounded-lg px-2 py-1.5 hover:bg-muted/50"
				>
					<Checkbox
						checked={value[option.key]}
						onCheckedChange={(checked) =>
							onChange({ ...value, [option.key]: checked })
						}
					/>
					{option.label}
				</label>
			))}
			<p className="text-muted-foreground">
				Project paths, credentials, and runtime data are never exported.
			</p>
		</fieldset>
	);
}

function ImportNameConflict({
	name,
	onNameChange,
	onImportCopy,
	onCancel,
}: {
	name: string;
	onNameChange: (name: string) => void;
	onImportCopy: () => void;
	onCancel: () => void;
}) {
	return (
		<section
			className="space-y-3 rounded-xl bg-destructive/5 p-4"
			aria-label="Import name conflict"
		>
			<p className="text-sm">
				A context named <strong>{name}</strong> already exists. Importing will
				not replace it.
			</p>
			<ContextField
				label="New context name"
				value={name}
				onChange={onNameChange}
			/>
			<div className="flex gap-2">
				<Button type="button" variant="outline" onClick={onImportCopy}>
					Import as copy
				</Button>
				<Button type="button" variant="ghost" onClick={onCancel}>
					Cancel import
				</Button>
			</div>
		</section>
	);
}

function normalizeContextName(name: string) {
	return name.trim().toLocaleLowerCase();
}

function nextCopyName(name: string, contextNames: string[]) {
	const base = name.trim() || "Imported context";
	const names = new Set(contextNames.map(normalizeContextName));
	for (let suffix = 1; ; suffix += 1) {
		const candidate = suffix === 1 ? `${base} copy` : `${base} copy ${suffix}`;
		if (!names.has(normalizeContextName(candidate))) return candidate;
	}
}

function Detail({ label, value }: { label: string; value: string }) {
	return (
		<div>
			<p className="text-xs uppercase text-muted-foreground">{label}</p>
			<p className="mt-1 break-all">{value}</p>
		</div>
	);
}
