import { MiniMap, useOnViewportChange, useReactFlow, useStore, useStoreApi } from "@xyflow/react";
import { List, Maximize, Minus, Plus } from "lucide-react";
import type { RefObject } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { familyNames, shapeOf } from "../../../model/graph";
import type { Family } from "../../../model/graph";
import type { FlowNode } from "./flow";
import { fittedViewport, steppedZoom } from "./viewportFit";
import { cx } from "../../cx";
import styles from "./GraphControls.module.css";
import vendor from "./vendor.module.css";

function minimapClass(node: FlowNode): string {
  if (node.type === "frame" || node.type === "frameTitle") return "minimap-frame";
  return node.selected === true ? "minimap-selected" : "";
}

const families = Object.keys(familyNames) as Family[];

function ReadingKeys() {
  return (
    <ul className={styles.legendList} aria-label="Readings on a node">
      <li>
        <span className={styles.glyph} aria-hidden="true">
          ·
        </span>
        Not reported by this run
      </li>
      <li>
        <span className={styles.glyph} aria-hidden="true">
          –
        </span>
        No reading yet at this time
      </li>
    </ul>
  );
}

function Legend({ review }: { readonly review: boolean }) {
  return (
    <div className={styles.legend} id="port-types" role="group" aria-labelledby="port-types-title">
      <p className={styles.legendTitle} id="port-types-title">
        Wire colour and pin shape by port type
      </p>
      <p className={styles.direction}>
        <span aria-hidden="true">→</span> Model connections carry values from outputs to inputs; arrows point to the
        receiving endpoint.
      </p>
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
        <li>
          <span className={cx(styles.swatch, "family-none")} aria-hidden="true" />
          Filled: connected
        </li>
        <li>
          <span className={cx(styles.swatch, styles.hollow, "family-none")} aria-hidden="true" />
          Hollow: not connected
        </li>
      </ul>
      {review && <ReadingKeys />}
    </div>
  );
}

function useLegendDismissal(open: boolean, anchor: RefObject<HTMLDivElement | null>, close: () => void) {
  useEffect(() => {
    if (!open) return undefined;
    const onKey = (event: globalThis.KeyboardEvent) => {
      if (event.key !== "Escape") return;
      event.stopPropagation();
      close();
      anchor.current?.querySelector("button")?.focus();
    };
    const onPointer = (event: PointerEvent) => {
      if (event.target instanceof Node && anchor.current?.contains(event.target) !== true) close();
    };
    document.addEventListener("keydown", onKey, true);
    document.addEventListener("pointerdown", onPointer, true);
    return () => {
      document.removeEventListener("keydown", onKey, true);
      document.removeEventListener("pointerdown", onPointer, true);
    };
  }, [open, anchor, close]);
}

function LegendButton({ review }: { readonly review: boolean }) {
  const [open, setOpen] = useState(false);
  const anchor = useRef<HTMLDivElement>(null);
  const close = useCallback(() => {
    setOpen(false);
  }, []);
  useLegendDismissal(open, anchor, close);
  return (
    <div className={styles.legendAnchor} ref={anchor}>
      <button
        type="button"
        className={styles.legendButton}
        aria-expanded={open}
        aria-controls="port-types"
        title="What the wire colours and pin shapes mean"
        onClick={() => {
          setOpen((value) => !value);
        }}
      >
        <List size={14} aria-hidden="true" />
        Legend
      </button>
      {open && <Legend review={review} />}
    </div>
  );
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
      fit: () => void flow.setViewport(fittedViewport(flow, store.getState(), 0.05), { duration: motion(120) }),
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

export function ZoomBar({ review = false }: { readonly review?: boolean }) {
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
      <LegendButton review={review} />
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
