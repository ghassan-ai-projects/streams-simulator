import { render } from "@testing-library/react";
import { NavigationStore } from "../app/navigationStore";
import { PlantMaker } from "../app/plantMaker";
import { WorkspaceStore } from "../app/workspaceStore";
import type { Badge } from "../model/report";
import { HomeScreen } from "../ui/home/HomeScreen";
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
  const store = new WorkspaceStore(bridge);
  await store.reload();
  return render(
    <HomeScreen workspace={store} navigation={nav.store} maker={new PlantMaker(bridge)} badgeFor={badgeFor} />,
  );
}
