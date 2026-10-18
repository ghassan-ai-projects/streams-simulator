import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { browserClock, hashLocation, localPreferences, mediaSystemTheme } from "./adapters/browser";
import { bridgeAdapter, browserTransport } from "./adapters/bridge";
import { createDeviceDesk, createDeviceSession } from "./app/devices";
import { createEditorSession } from "./app/editor";
import { createLaunchStore } from "./app/launch";
import { createPlantMaker, createPlantSession } from "./app/plant";
import { createReviewSession } from "./app/review";
import { createSession } from "./app/session";
import { createNavigationStore, createPanelStore, createThemeStore } from "./app/shell";
import { createWorkspaceStore } from "./app/workspace";
import { App } from "./ui/shell";
import "./ui/theme/tokens.css";
import "./ui/theme/base.css";
import "./ui/theme/families.css";

const root = document.getElementById("root");
if (root) {
  const preferences = localPreferences(window.localStorage);
  const themes = createThemeStore(preferences, mediaSystemTheme(window.matchMedia("(prefers-color-scheme: dark)")));
  const adapters = bridgeAdapter(browserTransport());
  const time = browserClock();
  const panels = createPanelStore(preferences);
  const navigation = createNavigationStore(hashLocation(window));
  const plants = createPlantSession(adapters.plant);
  const editor = createEditorSession(adapters.editor);
  const review = createReviewSession(adapters.review, time, time);
  createRoot(root).render(
    <StrictMode>
      <App
        stores={{
          themes,
          session: createSession(adapters.session),
          workspace: createWorkspaceStore(adapters.workspace),
          navigation,
          plants,
          review,
          launch: createLaunchStore(adapters.launch, time, time),
          scheduler: time,
          editor,
          devices: createDeviceSession(adapters.devices, adapters.workspace.watch),
          desk: createDeviceDesk(adapters.devices),
          panels,
          maker: createPlantMaker(adapters.plant),
          showDetails: () => {
            panels.show("right");
          },
        }}
      />
    </StrictMode>,
  );
}
