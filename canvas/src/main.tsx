import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { hashLocation, localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeAdapter, browserTransport } from "./adapters/bridge";
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
  const adapters = bridgeAdapter(browserTransport());
  const time = browserClock();
  const panels = new PanelStore(localPreferences(window.localStorage));
  const navigation = new NavigationStore(hashLocation(window));
  const plants = createPlantModule(adapters.plant).createSession();
  const editor = createEditorModule(adapters.editor).createSession();
  const review = createReviewModule(adapters.review, time, time).createSession();
  createRoot(root).render(
    <StrictMode>
      <App
        stores={{
          themes,
          session: new SessionStore(adapters.session),
          workspace: new WorkspaceStore(adapters.workspace),
          navigation,
          plants,
          review,
          launch: new LaunchStore(adapters.launch, time, time),
          scheduler: time,
          editor,
          panels,
          maker: new PlantMaker(adapters.plant),
          showDetails: () => {
            panels.show("right");
          },
        }}
      />
    </StrictMode>,
  );
}
