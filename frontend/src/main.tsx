import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import App from "./App";
import { Toaster } from "./components/ui/sonner";
import { ThemeProvider } from "next-themes";

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
	<React.StrictMode>
		<ThemeProvider attribute="class" defaultTheme="system" enableSystem>
			<App />
			<Toaster position="bottom-right" closeButton />
		</ThemeProvider>
	</React.StrictMode>,
);
