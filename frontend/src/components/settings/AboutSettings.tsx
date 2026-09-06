import { useEffect, useState } from "react";

import type { AboutState } from "../../lib/devctx-api";
import { devContextApi } from "../../lib/devctx-api.js";

function AboutSettings() {
	const [about, setAbout] = useState<AboutState>();

	useEffect(() => {
		void devContextApi.getAbout().then((result) => {
			if (result.ok) setAbout(result.data);
		});
	}, []);

	if (!about) return null;

	return (
		<section className="border-b border-border pb-6" aria-labelledby="settings-about">
			<h3 id="settings-about" className="font-semibold">About</h3>
			<p className="mt-1 text-sm text-muted-foreground">Dev Context {about.version}</p>
			<dl className="mt-4 grid gap-2 text-sm sm:grid-cols-2">
				<div><dt className="text-muted-foreground">License</dt><dd>{about.license}</dd></div>
				<div><dt className="text-muted-foreground">Build</dt><dd>{about.buildDate}</dd></div>
			</dl>
			<div className="mt-4 flex flex-wrap gap-4 text-sm">
				<a className="underline" href={about.documentationUrl}>Documentation</a>
				<a className="underline" href={about.repositoryUrl}>Repository</a>
			</div>
		</section>
	);
}

export { AboutSettings };
