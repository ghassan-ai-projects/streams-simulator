import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { browserClock, hashLocation, localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeAdapter, browserTransport } from "./adapters/bridge";
import { createEditorModule } from "./app/editor";
import { LaunchStore } from "./app/launch";
import { PlantMaker, createPlantModule } from "./app/plant";
import { createReviewModule } from "./app/review";
import { createSession } from "./app/session";
import { NavigationStore, PanelStore, ThemeStore } from "./app/shell";
import { WorkspaceStore } from "./app/workspace";
import { App } from "./ui/shell";
import "./ui/theme/tokens.css";
import "./ui/theme/base.css";
import "./ui/theme/families.css";

const root = document.getElementById("root");
if (root) {
  const preferences = localPreferences(window.localStorage);
  const themes = new ThemeStore(preferences, mediaSystemTheme(window.matchMedia("(prefers-color-scheme: dark)")));
  const adapters = bridgeAdapter(browserTransport());
  const time = browserClock();
  const panels = new PanelStore(preferences);
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
