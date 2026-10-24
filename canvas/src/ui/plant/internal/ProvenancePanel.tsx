import { Fingerprint, X } from "lucide-react";
import type { PlantMaker } from "../../../app/plant";
import type { ProvenanceSession } from "../../../app/workspace";
import type { NavigationStore } from "../../../app/shell";
import { plantOf, plantRoute } from "../../../model/workspace";
import { cx, panel, shared, useStore } from "../../primitives";
import { AssumptionsPanel } from "./AssumptionsPanel";
import type { Drawer } from "./drawer";
import { ExampleCopy } from "./ExampleCopy";
import { LineagePanel } from "./LineagePanel";
import { RouteWhy } from "./RouteNotice";

export interface ProvenancePanelProps {
  readonly provenance: ProvenanceSession;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly path: string;
  readonly title: string;
  readonly devices: readonly string[];
  readonly taken: readonly string[];
  readonly drawer: Drawer;
  readonly onClose: () => void;
}

const titles: Readonly<Record<Drawer, string>> = {
  lineage: "Lineage",
  assumptions: "Model assumptions",
  copy: "Editable copy",
};

function Body(props: ProvenancePanelProps) {
  const { provenance, path, drawer } = props;
  const lineage = useStore(provenance.lineage);
  const plant = plantOf(lineage, path);
  if (drawer === "assumptions") {
    return <AssumptionsPanel session={provenance.assumptions} path={path} twin={plant?.twin?.file ?? null} />;
  }
  if (plant === null || lineage === null) return null;
  if (drawer === "copy") {
    const { maker, navigation, taken, title } = props;
    return (
      <ExampleCopy
        example={plant.id}
        title={title}
        taken={taken}
        maker={maker}
        navigation={navigation}
        label="Create the copy"
      />
    );
  }
  const route = plantRoute(plant, props.devices);
  return (
    <>
      {route !== null && <RouteWhy route={route} />}
      <LineagePanel plant={plant} unreadable={lineage.unreadable} navigation={props.navigation} />
    </>
  );
}

export function ProvenancePanel(props: ProvenancePanelProps) {
  return (
    <aside className={cx(panel.panel, shared.scrolls)} aria-labelledby="provenance-title">
      <header className={panel.head}>
        <Fingerprint size={16} aria-hidden="true" />
        <h2 id="provenance-title">{titles[props.drawer]}</h2>
        <button type="button" className={panel.icon} aria-label="Close the provenance panel" onClick={props.onClose}>
          <X size={16} aria-hidden="true" />
        </button>
      </header>
      <div className={panel.body}>
        <Body {...props} />
      </div>
    </aside>
  );
}
