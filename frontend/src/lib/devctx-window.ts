export interface DevContextWindow {
	closeSelector(): Promise<void>;
}

export interface RuntimeBindings {
	quit(): Promise<void> | void;
}

export function createDevContextWindow(
	bindings: RuntimeBindings = generatedRuntimeBindings,
): DevContextWindow {
	return {
		async closeSelector() {
			await bindings.quit();
		},
	};
}

const generatedRuntimeBindings: RuntimeBindings = {
	async quit() {
		const runtime = await import("../../wailsjs/runtime/runtime");
		runtime.Quit();
	},
};

export const devContextWindow = createDevContextWindow();
