import { render } from "@testing-library/react";
import { createDeviceDesk } from "../app/devices";
import { createNavigationStore } from "../app/shell";
import { createPlantMaker } from "../app/plant";
import { createWorkspaceStore } from "../app/workspace";
import type { Badge } from "../model/review";
import { WorkspaceScreen } from "../ui/workspace";
import type { FakeBridge } from "./fakeBridge";
import { memoryLocation } from "./location";

export function navigation() {
  const location = memoryLocation();
  return { location, store: createNavigationStore(location) };
}

export async function show(
  bridge: FakeBridge,
  nav = navigation(),
  badgeFor: (worldId: string) => Badge | null = () => null,
) {
  const store = createWorkspaceStore(bridge.workspace);
  await store.open();
  return render(
    <WorkspaceScreen
      workspace={store}
      navigation={nav.store}
      maker={createPlantMaker(bridge.plant)}
      desk={createDeviceDesk(bridge.devices)}
      badgeFor={badgeFor}
    />,
  );
}
