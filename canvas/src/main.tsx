import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { hashLocation, localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeClient, browserTransport } from "./adapters/bridgeClient";
import { browserClock } from "./adapters/clock";
import { NavigationStore } from "./app/navigationStore";
import { PanelStore } from "./app/panelStore";
import { PlantMaker } from "./app/plantMaker";
import { LaunchStore } from "./app/launchStore";
import { createEditorModule, createPlantModule, createReviewModule } from "./app/routeModules";
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
  const time = browserClock();
  const panels = new PanelStore(localPreferences(window.localStorage));
  const navigation = new NavigationStore(hashLocation(window));
  const plants = createPlantModule(bridge).createSession();
  const editor = createEditorModule(bridge).createSession();
  const review = createReviewModule(bridge, time, time).createSession();
  createRoot(root).render(
    <StrictMode>
      <App
        stores={{
          themes,
          session: new SessionStore(bridge),
          workspace: new WorkspaceStore(bridge),
          navigation,
          plants,
          review,
          launch: new LaunchStore(bridge, time, time),
          scheduler: time,
          editor,
          panels,
          maker: new PlantMaker(bridge),
          showDetails: () => {
            panels.show("right");
          },
        }}
      />
    </StrictMode>,
  );
}
