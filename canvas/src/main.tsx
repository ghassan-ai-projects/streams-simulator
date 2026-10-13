import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeClient, browserTransport } from "./adapters/bridgeClient";
import { SessionStore } from "./app/sessionStore";
import { ThemeStore } from "./app/themeStore";
import { WorkspaceStore } from "./app/workspaceStore";
import { App } from "./ui/App";
import "./ui/theme/tokens.css";
import "./ui/theme/base.css";

const root = document.getElementById("root");
if (root) {
  const themes = new ThemeStore(
    localPreferences(window.localStorage),
    mediaSystemTheme(window.matchMedia("(prefers-color-scheme: dark)")),
  );
  const bridge = bridgeClient(browserTransport());
  createRoot(root).render(
    <StrictMode>
      <App themes={themes} session={new SessionStore(bridge)} workspace={new WorkspaceStore(bridge)} />
    </StrictMode>,
  );
}
