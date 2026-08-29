import type { Ref } from "react";
import type { ContextState, LaunchState } from "../../lib/devctx-api";
import { ContextCard } from "./ContextCard.js";
import { contextRecommendation } from "./recommendation.js";
import type { ContextNavigationDirection } from "./selection-state.js";

const contextSearchThreshold = 6;

interface ContextChoiceListProps {
	launchState: LaunchState;
	selectedContextId?: string;
	rovingContextId?: string;
	launchPending: boolean;
	keyboardLaunchAvailable: boolean;
	search: string;
	onSearchChange: (search: string) => void;
	buttonRef: (contextId: string) => Ref<HTMLButtonElement>;
	onSelect: (contextId: string) => void;
	onNavigate: (
		contextId: string,
		direction: ContextNavigationDirection,
	) => void;
	onLaunch: () => void;
	onProviderSetup: (contextId: string, providerId: string) => void;
}

function ContextChoiceList({
	launchState,
	selectedContextId,
	rovingContextId,
	launchPending,
	keyboardLaunchAvailable,
	search,
	onSearchChange,
	buttonRef,
	onSelect,
	onNavigate,
	onLaunch,
	onProviderSetup,
}: ContextChoiceListProps) {
	const contexts = filterContexts(launchState.contexts, search);
	const groups = groupContexts(
		contexts.map((context) => ({
			context,
			recommendation: contextRecommendation(launchState, context),
		})),
	);

	return (
		<div className="space-y-5">
			{shouldShowContextSearch(launchState.contexts) ? (
				<label className="block">
					<span className="sr-only">Search contexts</span>
					<input
						type="search"
						value={search}
						onChange={(event) => onSearchChange(event.target.value)}
						placeholder="Search contexts"
						className="w-full border border-input bg-background px-3 py-2 text-sm"
					/>
				</label>
			) : null}
			{groups.length === 0 ? (
				<p className="text-sm text-muted-foreground">
					No contexts match your search.
				</p>
			) : (
				groups.map((group) => (
					<section
						key={group.label}
						aria-labelledby={`context-group-${group.label}`}
					>
						<h3
							id={`context-group-${group.label}`}
							className="mb-3 text-sm font-semibold"
						>
							{group.label}
						</h3>
						<div className="grid gap-4 sm:grid-cols-2">
							{group.contexts.map(({ context, recommendation }) => (
								<ContextCard
									key={context.id}
									context={context}
									compact
									selected={selectedContextId === context.id}
									recommendation={recommendation}
									disabled={launchPending}
									tabIndex={rovingContextId === context.id ? 0 : -1}
									buttonRef={buttonRef(context.id)}
									onSelect={onSelect}
									onNavigate={onNavigate}
									onLaunchSelected={
										keyboardLaunchAvailable ? onLaunch : undefined
									}
									onProviderSetup={onProviderSetup}
								/>
							))}
						</div>
					</section>
				))
			)}
		</div>
	);
}

function shouldShowContextSearch(contexts: ContextState[]): boolean {
	return contexts.length >= contextSearchThreshold;
}

function filterContexts(
	contexts: ContextState[],
	search: string,
): ContextState[] {
	const query = search.trim().toLocaleLowerCase();
	if (query === "") {
		return contexts;
	}
	return contexts.filter((context) =>
		[context.name, context.description]
			.filter((value): value is string => value !== undefined)
			.some((value) => value.toLocaleLowerCase().includes(query)),
	);
}

function groupContexts(
	contexts: Array<{
		context: ContextState;
		recommendation: ReturnType<typeof contextRecommendation>;
	}>,
): Array<{
	label: "Remembered context" | "Other contexts" | "Contexts";
	contexts: typeof contexts;
}> {
	const remembered = contexts.find(
		({ recommendation }) => recommendation?.category === "remembered",
	);
	if (remembered === undefined) {
		return contexts.length === 0 ? [] : [{ label: "Contexts", contexts }];
	}

	const otherContexts = contexts.filter(
		({ context }) => context.id !== remembered.context.id,
	);
	return [
		{ label: "Remembered context", contexts: [remembered] },
		...(otherContexts.length === 0
			? []
			: [{ label: "Other contexts" as const, contexts: otherContexts }]),
	];
}

export {
	ContextChoiceList,
	filterContexts,
	groupContexts,
	shouldShowContextSearch,
};
