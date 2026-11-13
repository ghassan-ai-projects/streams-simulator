import { MiniMap, useOnViewportChange, useReactFlow, useStore, useStoreApi } from "@xyflow/react";
import { Maximize, Minus, Plus } from "lucide-react";
import { useMemo, useState } from "react";
import type { FlowNode } from "./flow";
import { fittedViewport, lowestZoom, steppedZoom } from "./viewportFit";
import { cx } from "../../primitives";
import { LegendButton } from "./Legend";
import type { WiresControl } from "./useWires";
import { WiresToggle } from "./WiresToggle";
import styles from "./GraphControls.module.css";
import vendor from "./vendor.module.css";

function minimapClass(node: FlowNode): string {
  if (node.type === "frame" || node.type === "frameTitle") return cx(styles.minimapNode, styles.minimapFrame);
  return cx(styles.minimapNode, node.selected === true && styles.minimapSelected);
}

function motion(ms: number): number {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches ? 0 : ms;
}

export function useZoomActions() {
  const flow = useReactFlow<FlowNode>();
  const store = useStoreApi<FlowNode>();
  return useMemo(
    () => ({
      in: () => void flow.setViewport(steppedZoom(flow, store.getState(), 1.2), { duration: motion(120) }),
      out: () => void flow.setViewport(steppedZoom(flow, store.getState(), 1 / 1.2), { duration: motion(120) }),
      fit: () => void flow.setViewport(fittedViewport(flow, store.getState(), lowestZoom), { duration: motion(120) }),
    }),
    [flow, store],
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

export function ZoomBar({
  review = false,
  wires = null,
  linking = false,
}: {
  readonly review?: boolean;
  readonly wires?: WiresControl | null;
  readonly linking?: boolean;
}) {
  const zoomTo = useZoomActions();
  const zoom = useStore((state) => state.transform[2]);
  const settled = useSettledZoom();
  return (
    <div className={styles.zoom} role="group" aria-label="Zoom">
      <button type="button" className={styles.button} aria-label="Zoom out" title="Zoom out" onClick={zoomTo.out}>
        <Minus size={14} aria-hidden="true" />
      </button>
      <span className={styles.percent} aria-hidden="true">{`${String(Math.round(zoom * 100))} %`}</span>
      <span className="sr-only" aria-live="polite">
        {settled}
      </span>
      <button type="button" className={styles.button} aria-label="Zoom in" title="Zoom in" onClick={zoomTo.in}>
        <Plus size={14} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={styles.button}
        aria-label="Fit the plant in view"
        title="Fit the plant in view"
        onClick={zoomTo.fit}
      >
        <Maximize size={14} aria-hidden="true" />
      </button>
      {wires !== null && <WiresToggle control={wires} linking={linking} />}
      <LegendButton review={review} grouped={wires?.value === "grouped"} />
    </div>
  );
}

export function Navigator() {
  return (
    <section className={styles.navigator} aria-label="Navigator">
      <h2 className={styles.title}>Navigator</h2>
      <MiniMap
        pannable
        zoomable
        className={cx(styles.minimap, vendor.minimap)}
        ariaLabel="Navigator map"
        nodeClassName={minimapClass}
      />
    </section>
  );
}
