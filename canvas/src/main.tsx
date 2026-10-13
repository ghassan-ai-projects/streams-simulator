import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { localPreferences, mediaSystemTheme } from "./adapters/browser";
import { ThemeStore } from "./app/themeStore";
import { App } from "./ui/App";
import "./ui/theme/tokens.css";
import "./ui/theme/base.css";

const root = document.getElementById("root");
if (root) {
  const themes = new ThemeStore(
    localPreferences(window.localStorage),
    mediaSystemTheme(window.matchMedia("(prefers-color-scheme: dark)")),
  );
  createRoot(root).render(
    <StrictMode>
      <App themes={themes} />
    </StrictMode>,
  );
}
