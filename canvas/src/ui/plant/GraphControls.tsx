import { MiniMap, useReactFlow, useStore } from "@xyflow/react";
import { List, Maximize, Minus, Plus } from "lucide-react";
import { useState } from "react";
import { familyNames, shapeOf } from "../../model/ports";
import type { Family } from "../../model/ports";
import type { FlowNode } from "./flow";
import { cx } from "../cx";
import styles from "./GraphControls.module.css";

function minimapClass(node: FlowNode): string {
  if (node.type === "frame") return "minimap-frame";
  return node.selected === true ? "minimap-selected" : "";
}

const families = Object.keys(familyNames) as Family[];

function Legend() {
  return (
    <div className={styles.legend} role="dialog" aria-label="Port types">
      <ul className={styles.legendList}>
        {families.map((family) => (
          <li key={family}>
            <span className={cx(styles.swatch, styles[shapeOf(family)], `family-${family}`)} aria-hidden="true" />
            {familyNames[family]}
          </li>
        ))}
        <li>
          <span className={cx(styles.swatch, styles.unknown, "family-none")} aria-hidden="true" />
          Port not recorded
        </li>
      </ul>
    </div>
  );
}

function LegendButton() {
  const [open, setOpen] = useState(false);
  return (
    <div className={styles.legendAnchor}>
      <button
        type="button"
        className={styles.button}
        aria-expanded={open}
        aria-label="Port types"
        onClick={() => {
          setOpen((value) => !value);
        }}
      >
        <List size={14} aria-hidden="true" />
      </button>
      {open && <Legend />}
    </div>
  );
}

export function ZoomBar() {
  const flow = useReactFlow<FlowNode>();
  const zoom = useStore((state) => state.transform[2]);
  return (
    <div className={styles.zoom} role="group" aria-label="Zoom">
      <button
        type="button"
        className={styles.button}
        aria-label="Zoom out"
        onClick={() => void flow.zoomOut({ duration: 120 })}
      >
        <Minus size={14} aria-hidden="true" />
      </button>
      <span className={styles.percent} aria-live="polite">{`${String(Math.round(zoom * 100))} %`}</span>
      <button
        type="button"
        className={styles.button}
        aria-label="Zoom in"
        onClick={() => void flow.zoomIn({ duration: 120 })}
      >
        <Plus size={14} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={styles.button}
        aria-label="Fit the plant in view"
        onClick={() => void flow.fitView({ padding: 0.04, duration: 120 })}
      >
        <Maximize size={14} aria-hidden="true" />
      </button>
      <LegendButton />
    </div>
  );
}

export function Navigator() {
  return (
    <section className={styles.navigator} aria-label="Navigator">
      <h2 className={styles.title}>Navigator</h2>
      <MiniMap pannable zoomable className={styles.minimap} ariaLabel="Navigator map" nodeClassName={minimapClass} />
    </section>
  );
}
