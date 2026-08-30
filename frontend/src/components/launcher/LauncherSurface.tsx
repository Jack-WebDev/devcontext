import type { ReactNode } from "react";

interface LauncherSurfaceProps {
	projectPath: string;
	children: ReactNode;
}

function LauncherSurface({ projectPath, children }: LauncherSurfaceProps) {
	return (
		<main
			className="launcher-surface flex min-h-screen items-center justify-center bg-background"
			data-launcher-flow
		>
			<section
				aria-labelledby="launcher-heading"
				className="launcher-container border border-border bg-card shadow-sm"
			>
				<p className="text-label text-secondary">Dev Context</p>
				<h1 id="launcher-heading" className="text-launcher-title mt-2">
					Open project
				</h1>
				<p
					className="text-technical text-secondary mt-4 break-all"
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
