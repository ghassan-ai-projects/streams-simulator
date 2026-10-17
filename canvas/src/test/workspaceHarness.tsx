import { render } from "@testing-library/react";
import { NavigationStore } from "../app/shell";
import { PlantMaker } from "../app/plant";
import { WorkspaceStore } from "../app/workspace";
import type { Badge } from "../model/review";
import { WorkspaceScreen } from "../ui/workspace";
import type { FakeBridge } from "./fakeBridge";
import { memoryLocation } from "./location";

export function navigation() {
  const location = memoryLocation();
  return { location, store: new NavigationStore(location) };
}

export async function show(
  bridge: FakeBridge,
  nav = navigation(),
  badgeFor: (worldId: string) => Badge | null = () => null,
) {
  const store = new WorkspaceStore(bridge.workspace);
  await store.open();
  return render(
    <WorkspaceScreen
      workspace={store}
      navigation={nav.store}
      maker={new PlantMaker(bridge.plant)}
      badgeFor={badgeFor}
    />,
  );
}
