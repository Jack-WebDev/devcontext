import { useEffect, useState } from "react";
import type {
	ApiResult,
	ContextListItem,
	ContextTemplateState,
	CreateContextRequest,
	CreateContextResult,
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
import { useContextCreation } from "./context-creation";

export { ContextDetailsDrawer } from "./ContextDetailsDrawer";

type CustomContextRequest = CreateContextRequest & {
	name: string;
	description: string;
	icon: string;
	accent: string;
	toolId: string;
	enabledProviderIds: string[];
};

export function CreateContextDialog({
	contexts,
	onClose,
	create,
	loadTemplates,
}: {
	contexts: ContextListItem[];
	onClose: () => void;
	create: (
		request: CreateContextRequest,
	) => Promise<ApiResult<CreateContextResult>>;
	loadTemplates: () => Promise<
		ApiResult<{ templates: ContextTemplateState[] }>
	>;
}) {
	const [name, setName] = useState("");
	const [description, setDescription] = useState("");
	const [icon, setIcon] = useState("");
	const [accent, setAccent] = useState("custom");
	const [toolId, setToolID] = useState(contexts[0]?.context.tool.id ?? "");
	const [providers, setProviders] = useState<string[]>([]);
	const [templates, setTemplates] = useState<ContextTemplateState[]>([]);
	const [templateID, setTemplateID] = useState("custom");
	const contextCreation = useContextCreation(create);
	const options = contexts[0]?.context.availableTools ?? [];
	const providerOptions = contexts
		.flatMap((c) => c.context.providers)
		.filter((p, i, all) => all.findIndex((x) => x.id === p.id) === i);
	useEffect(() => {
		void loadTemplates().then((result) => {
			if (result.ok) setTemplates(result.data.templates);
		});
	}, [loadTemplates]);
	function selectTemplate(id: string) {
		setTemplateID(id);
		const template = templates.find((item) => item.id === id);
		if (!template) return;
		setName(template.name);
		setDescription(template.description);
		setIcon(template.icon ?? "");
		setAccent(template.accent);
	}
	async function submit() {
		const request: CustomContextRequest = {
			templateId: templateID,
			name,
			description,
			icon,
			accent,
			toolId,
			enabledProviderIds: providers,
		};
		const result = await contextCreation.create(request);
		if (result === undefined || !result.ok) {
			return;
		}
		onClose();
	}
	return (
		<Sheet open onOpenChange={(open) => !open && onClose()}>
			<SheetContent>
				<SheetHeader>
					<SheetTitle>New context</SheetTitle>
					<SheetDescription>
						Create an isolated development identity.
					</SheetDescription>
				</SheetHeader>
				<div className="space-y-4 px-8 pb-8">
					<label className="block text-sm">
						Start from a template
						<select
							className="mt-1 w-full border p-2"
							value={templateID}
							onChange={(e) => selectTemplate(e.target.value)}
						>
							{templates.map((template) => (
								<option key={template.id} value={template.id}>
									{template.name}
								</option>
							))}
						</select>
					</label>
					<ContextField label="Name" value={name} onChange={setName} />
					<ContextField
						label="Description"
						value={description}
						onChange={setDescription}
					/>
					<ContextField label="Icon" value={icon} onChange={setIcon} />
					<label className="block text-sm">
						Accent
						<select
							className="mt-1 w-full border p-2"
							value={accent}
							onChange={(e) => setAccent(e.target.value)}
						>
							{["sage", "slate-blue", "amber", "custom"].map((value) => (
								<option key={value}>{value}</option>
							))}
						</select>
					</label>
					{options.length > 0 ? (
						<label className="block text-sm">
							Coding tool
							<select
								className="mt-1 w-full border p-2"
								value={toolId}
								onChange={(e) => setToolID(e.target.value)}
							>
								{options.map((tool) => (
									<option key={tool.id} value={tool.id}>
										{tool.name}
									</option>
								))}
							</select>
						</label>
					) : (
						<p className="text-sm text-muted-foreground">
							No coding tools detected. You can create this context now and
							configure a coding tool later.
						</p>
					)}
					<fieldset>
						<legend className="text-sm">Providers</legend>
						{providerOptions.map((provider) => (
							<label key={provider.id} className="block">
								<input
									type="checkbox"
									checked={providers.includes(provider.id)}
									onChange={() =>
										setProviders((current) =>
											current.includes(provider.id)
												? current.filter((id) => id !== provider.id)
												: [...current, provider.id],
										)
									}
								/>{" "}
								{provider.name}
							</label>
						))}
					</fieldset>
					{contextCreation.error ? (
						<p className="text-destructive">{contextCreation.error.message}</p>
					) : null}
					<Button
						type="button"
						disabled={contextCreation.pending || !name}
						onClick={() => void submit()}
					>
						{contextCreation.pending ? "Creating..." : "Create context"}
					</Button>
				</div>
			</SheetContent>
		</Sheet>
	);
}
