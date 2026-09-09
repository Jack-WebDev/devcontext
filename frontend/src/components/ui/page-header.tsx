import type { ReactNode } from "react";

import { cn } from "../../lib/utils.js";

interface PageHeaderProps {
	id: string;
	title: string;
	eyebrow?: string;
	description?: string;
	actions?: ReactNode;
	className?: string;
}

function PageHeader({
	id,
	title,
	eyebrow,
	description,
	actions,
	className,
}: PageHeaderProps) {
	return (
		<header className={cn("page-header", className)}>
			<div className="page-header-copy">
				{eyebrow ? <p className="page-header-eyebrow">{eyebrow}</p> : null}
				<h2 id={id} className="text-page-title">
					{title}
				</h2>
				{description ? (
					<p className="page-header-description">{description}</p>
				) : null}
			</div>
			{actions ? (
				<div className="control-cluster shrink-0">{actions}</div>
			) : null}
		</header>
	);
}

export type { PageHeaderProps };
export { PageHeader };
