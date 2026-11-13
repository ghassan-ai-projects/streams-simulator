import { useStore } from "@xyflow/react";
import { List } from "lucide-react";
import type { RefObject } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Family } from "../../../model/plant";
import { familyNames, shapeOf } from "../../../model/plant";
import { cx } from "../../primitives";
import type { FlowNode } from "./flow";
import { useHiddenFamilies } from "./wireFilter";
import styles from "./GraphControls.module.css";

const families = Object.keys(familyNames) as Family[];

function familiesIn(nodes: readonly FlowNode[]): string {
  const used = new Set<string>();
  for (const node of nodes) {
    if (node.type !== "instance") continue;
    const { ins, outs } = node.data.view.pins;
    for (const pin of [...ins, ...outs]) if (pin.family !== null) used.add(pin.family);
  }
  return families.filter((family) => used.has(family)).join(" ");
}

function usePlantFamilies(): readonly Family[] {
  const used = useStore((state) => familiesIn(state.nodes as FlowNode[]));
  return useMemo(
    () => (used === "" ? families : families.filter((family) => used.split(" ").includes(family))),
    [used],
  );
}

function ReadingKeys() {
  return (
    <ul className={styles.legendList} aria-label="Readings on a node">
      <li>
        <span className={styles.glyph} aria-hidden="true">
          ·
        </span>
        No reading from this run for this row (its words say why)
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

function PinKeys() {
  return (
    <>
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
    </>
  );
}

function GroupedKey() {
  return (
    <p className={styles.direction}>
      <span aria-hidden="true">×3</span> Three links between the same two instances, drawn as one wire in the colour of
      the material or power it carries. A wire that carries several kinds is grey.
    </p>
  );
}

function FamilyKey({ family }: { readonly family: Family }) {
  const { hidden, toggle } = useHiddenFamilies();
  const shown = !hidden.has(family);
  return (
    <li>
      <button
        type="button"
        className={cx(styles.familyKey, !shown && styles.familyOff)}
        aria-pressed={shown}
        title={shown ? `Hide the ${familyNames[family]} wires` : `Show the ${familyNames[family]} wires`}
        onClick={() => {
          toggle(family);
        }}
      >
        <span className={cx(styles.swatch, styles[shapeOf(family)], `family-${family}`)} aria-hidden="true" />
        <span className={styles.familyName}>{familyNames[family]}</span>
        {!shown && <span className={styles.offWord}>hidden</span>}
      </button>
    </li>
  );
}

function Legend({ review, grouped }: { readonly review: boolean; readonly grouped: boolean }) {
  const shown = usePlantFamilies();
  return (
    <div className={styles.legend} id="port-types" role="group" aria-labelledby="port-types-title">
      <p className={styles.legendTitle} id="port-types-title">
        Wire colour and pin shape by port type
      </p>
      <p className={styles.direction}>
        <span aria-hidden="true">→</span> Model connections carry values from outputs to inputs; arrows point to the
        receiving endpoint. Select a type to hide or show its wires.
      </p>
      {grouped && <GroupedKey />}
      <ul className={styles.legendList}>
        {shown.map((family) => (
          <FamilyKey key={family} family={family} />
        ))}
        {!grouped && <PinKeys />}
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

export function LegendButton({ review, grouped }: { readonly review: boolean; readonly grouped: boolean }) {
  const [open, setOpen] = useState(false);
  const hiddenCount = useHiddenFamilies().hidden.size;
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
        {hiddenCount > 0 && <span className={styles.offWord}>{`· ${String(hiddenCount)} hidden`}</span>}
      </button>
      {open && <Legend review={review} grouped={grouped} />}
    </div>
  );
}
