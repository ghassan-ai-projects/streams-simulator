import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { hashLocation, localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeClient, browserTransport } from "./adapters/bridgeClient";
import { EditorStore } from "./app/editorStore";
import { NavigationStore } from "./app/navigationStore";
import { PlantStore } from "./app/plantStore";
import { LaunchStore } from "./app/launchStore";
import { ReviewStore } from "./app/reviewStore";
import { SessionStore } from "./app/sessionStore";
import { ThemeStore } from "./app/themeStore";
import { WorkspaceStore } from "./app/workspaceStore";
import { App } from "./ui/shell/App";
import "./ui/theme/tokens.css";
import "./ui/theme/base.css";
import "./ui/theme/families.css";

const root = document.getElementById("root");
if (root) {
  const themes = new ThemeStore(
    localPreferences(window.localStorage),
    mediaSystemTheme(window.matchMedia("(prefers-color-scheme: dark)")),
  );
  const bridge = bridgeClient(browserTransport());
  createRoot(root).render(
    <StrictMode>
      <App
        stores={{
          themes,
          session: new SessionStore(bridge),
          workspace: new WorkspaceStore(bridge),
          navigation: new NavigationStore(hashLocation(window)),
          plants: new PlantStore(bridge),
          review: new ReviewStore(bridge),
          launch: new LaunchStore(bridge),
          editor: new EditorStore(bridge),
        }}
      />
    </StrictMode>,
  );
}
