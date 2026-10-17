import { render } from "@testing-library/react";
import { NavigationStore } from "../app/shell";
import { PlantMaker } from "../app/plant";
import { WorkspaceStore } from "../app/workspace";
import type { Badge } from "../model/review";
import { HomeScreen } from "../ui/workspace";
import type { FakeBridge } from "./fakeBridge";

export function navigation() {
  const location = {
    hash: "#/",
    read() {
      return this.hash;
    },
    write(hash: string) {
      this.hash = hash;
    },
    subscribe: () => () => undefined,
  };
  return { location, store: new NavigationStore(location) };
}

export async function show(bridge: FakeBridge, nav = navigation(), badgeFor?: (worldId: string) => Badge | null) {
  const store = new WorkspaceStore(bridge.workspace);
  await store.reload();
  return render(
    <HomeScreen
      workspace={store}
      navigation={nav.store}
      maker={new PlantMaker(bridge.plant)}
      {...(badgeFor === undefined ? {} : { badgeFor })}
    />,
  );
}
