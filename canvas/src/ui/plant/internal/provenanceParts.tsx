import type { ReactNode } from "react";
import { useState } from "react";
import type { PlantMaker } from "../../../app/plant";
import type { ProvenanceSession } from "../../../app/workspace";
import type { NavigationStore } from "../../../app/shell";
import type { Workspace } from "../../../model/workspace";
import { plantsOf } from "../../../model/workspace";
import type { Drawer } from "./drawer";
import { ProvenanceBlock } from "./ProvenanceBlock";
import { ProvenancePanel } from "./ProvenancePanel";

export function useDrawer(onShowDetails: () => void) {
  const [drawer, setDrawer] = useState<Drawer | null>(null);
  const pick = (next: Drawer | null) => {
    setDrawer(next);
    if (next !== null) onShowDetails();
  };
  return {
    drawer,
    pick,
    close: () => {
      setDrawer(null);
    },
  };
}

interface PartsProps {
  readonly provenance: ProvenanceSession;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly path: string;
  readonly title: string;
  readonly workspace: Workspace | null;
  readonly drawer: ReturnType<typeof useDrawer>;
}

export function provenanceParts(props: PartsProps): { readonly block: ReactNode; readonly panel: ReactNode } {
  const { provenance, navigation, maker, path, title, workspace, drawer } = props;
  const devices = workspace?.kinds.map((kind) => kind.path) ?? [];
  const taken = workspace === null ? [] : plantsOf(workspace).map((plant) => plant.id);
  const block = (
    <ProvenanceBlock
      provenance={provenance}
      navigation={navigation}
      path={path}
      devices={devices}
      drawer={drawer.drawer}
      onPick={drawer.pick}
    />
  );
  if (drawer.drawer === null) return { block, panel: null };
  const panel = (
    <ProvenancePanel
      provenance={provenance}
      navigation={navigation}
      maker={maker}
      path={path}
      title={title}
      devices={devices}
      taken={taken}
      drawer={drawer.drawer}
      onClose={drawer.close}
    />
  );
  return { block, panel };
}
