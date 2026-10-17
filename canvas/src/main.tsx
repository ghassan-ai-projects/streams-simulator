import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { hashLocation, localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeAdapter, browserTransport } from "./adapters/bridge";
import { browserClock } from "./adapters/browser";
import { NavigationStore } from "./app/shell";
import { PanelStore } from "./app/shell";
import { PlantMaker } from "./app/plant";
import { LaunchStore } from "./app/launch";
import { createEditorModule } from "./app/editor";
import { createPlantModule } from "./app/plant";
import { createReviewModule } from "./app/review";
import { createSession } from "./app/session";
import { ThemeStore } from "./app/shell";
import { WorkspaceStore } from "./app/workspace";
import { App } from "./ui/shell";
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
          session: createSession(adapters.session),
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
