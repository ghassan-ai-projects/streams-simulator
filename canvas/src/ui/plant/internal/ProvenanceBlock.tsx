import type { ProvenanceSession } from "../../../app/workspace";
import type { NavigationStore } from "../../../app/shell";
import { plantOf, plantRoute } from "../../../model/workspace";
import { shared, useStore } from "../../primitives";
import type { Drawer } from "./drawer";
import { RouteNotice } from "./RouteNotice";
import styles from "./Provenance.module.css";

export interface BlockProps {
  readonly provenance: ProvenanceSession;
  readonly navigation: NavigationStore;
  readonly path: string;
  readonly devices: readonly string[];
  readonly drawer: Drawer | null;
  readonly onPick: (next: Drawer | null) => void;
}

interface ToggleProps {
  readonly drawer: Drawer | null;
  readonly name: "lineage" | "assumptions";
  readonly label: string;
  readonly unavailable?: boolean;
  readonly onPick: BlockProps["onPick"];
}

function Toggle({ drawer, name, label, unavailable = false, onPick }: ToggleProps) {
  const open = drawer === name;
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-expanded={open}
      aria-disabled={unavailable}
      title={unavailable ? "The bridge has not given the lineage of this workspace" : undefined}
      onClick={() => {
        if (!unavailable) onPick(open ? null : name);
      }}
    >
      {label}
    </button>
  );
}

export function ProvenanceBlock({ provenance, navigation, path, devices, drawer, onPick }: BlockProps) {
  const lineage = useStore(provenance.lineage);
  const plant = plantOf(lineage, path);
  const route = plantRoute(plant, devices);
  return (
    <section className={styles.block} aria-label="Provenance of this plant">
      {route !== null && (
        <RouteNotice
          route={route}
          provenance={provenance}
          navigation={navigation}
          onCopy={() => {
            onPick(drawer === "copy" ? null : "copy");
          }}
        />
      )}
      <span className={styles.toggles}>
        <Toggle drawer={drawer} name="lineage" label="Lineage" unavailable={plant === null} onPick={onPick} />
        <Toggle drawer={drawer} name="assumptions" label="Model assumptions" onPick={onPick} />
      </span>
    </section>
  );
}
