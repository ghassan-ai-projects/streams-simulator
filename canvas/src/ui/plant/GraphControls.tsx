import { MiniMap, useOnViewportChange, useReactFlow, useStore } from "@xyflow/react";
import { List, Maximize, Minus, Plus } from "lucide-react";
import type { KeyboardEvent } from "react";
import { useMemo, useRef, useState } from "react";
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
    <div className={styles.legend} id="port-types">
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
  const button = useRef<HTMLButtonElement>(null);
  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Escape" || !open) return;
    event.stopPropagation();
    setOpen(false);
    button.current?.focus();
  };
  return (
    <div className={styles.legendAnchor} onKeyDown={onKeyDown}>
      <button
        ref={button}
        type="button"
        className={styles.button}
        aria-expanded={open}
        aria-controls="port-types"
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

function motion(ms: number): number {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches ? 0 : ms;
}

export function useZoomActions() {
  const flow = useReactFlow<FlowNode>();
  return useMemo(
    () => ({
      in: () => void flow.zoomIn({ duration: motion(120) }),
      out: () => void flow.zoomOut({ duration: motion(120) }),
      fit: () => void flow.fitView({ padding: 0.04, duration: motion(120) }),
      centre: (x: number, y: number, zoom: number) => void flow.setCenter(x, y, { zoom, duration: motion(200) }),
    }),
    [flow],
  );
}

function useSettledZoom(): string {
  const [settled, setSettled] = useState("");
  useOnViewportChange({
    onEnd: (viewport) => {
      setSettled(`Zoom ${String(Math.round(viewport.zoom * 100))} %`);
    },
  });
  return settled;
}

export function ZoomBar() {
  const zoomTo = useZoomActions();
  const zoom = useStore((state) => state.transform[2]);
  const settled = useSettledZoom();
  return (
    <div className={styles.zoom} role="group" aria-label="Zoom">
      <button type="button" className={styles.button} aria-label="Zoom out" onClick={zoomTo.out}>
        <Minus size={14} aria-hidden="true" />
      </button>
      <span className={styles.percent} aria-hidden="true">{`${String(Math.round(zoom * 100))} %`}</span>
      <span className="sr-only" aria-live="polite">
        {settled}
      </span>
      <button type="button" className={styles.button} aria-label="Zoom in" onClick={zoomTo.in}>
        <Plus size={14} aria-hidden="true" />
      </button>
      <button type="button" className={styles.button} aria-label="Fit the plant in view" onClick={zoomTo.fit}>
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
