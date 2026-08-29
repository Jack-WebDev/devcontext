import type { ReactNode } from "react";

interface LauncherSurfaceProps {
	projectPath: string;
	children: ReactNode;
}

function LauncherSurface({ projectPath, children }: LauncherSurfaceProps) {
	return (
		<main
			className="flex min-h-screen items-center justify-center bg-background p-6"
			data-launcher-flow
		>
			<section
				aria-labelledby="launcher-heading"
				className="w-full max-w-5xl border border-border bg-card p-6 shadow-sm"
			>
				<p className="text-sm font-semibold text-muted-foreground">Dev Context</p>
				<h1 id="launcher-heading" className="mt-2 text-2xl font-semibold">
					Open project
				</h1>
				<p
					className="mt-4 break-all font-mono text-sm text-muted-foreground"
					data-launcher-project-path
				>
					{projectPath}
				</p>
				<div className="mt-6">{children}</div>
			</section>
		</main>
	);
}

export { LauncherSurface };
