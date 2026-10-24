import { CircleCheck, Radio, SlidersHorizontal, TriangleAlert } from "lucide-react";
import type { AgendaItem, AgendaLane } from "../../../model/launch";
import { panel } from "../../primitives";
import styles from "./PlanAgenda.module.css";

const icons = {
  fault: TriangleAlert,
  clear: CircleCheck,
  delivery: Radio,
  "delivery-end": Radio,
  call: SlidersHorizontal,
} as const;

const laneClass: Readonly<Record<AgendaLane, string | undefined>> = {
  fault: styles.fault,
  clear: styles.clear,
  delivery: styles.delivery,
  "delivery-end": styles.deliveryEnd,
  call: styles.call,
};

function Item({ item }: { readonly item: AgendaItem }) {
  const Icon = icons[item.lane];
  return (
    <li className={laneClass[item.lane]}>
      <button
        type="button"
        className={styles.item}
        aria-label={`${item.time} ${item.label}: ${item.text}. Edit it.`}
        onClick={() => {
          document.getElementById(item.editor)?.focus();
        }}
      >
        <span className={styles.time}>{item.time}</span>
        <span className={styles.what}>
          <span className={styles.lane}>
            <Icon size={12} aria-hidden="true" />
            {item.label}
          </span>
          <span className={styles.text}>{item.text}</span>
        </span>
      </button>
    </li>
  );
}

export function PlanAgenda({ items }: { readonly items: readonly AgendaItem[] }) {
  return (
    <section className={panel.group} aria-labelledby="launch-agenda-title">
      <h3 id="launch-agenda-title">Everything planned, in time order</h3>
      {items.length === 0 ? (
        <p className={panel.hint}>Nothing yet: add a fault, a delivery condition or an effector call above.</p>
      ) : (
        <ol className={styles.list} aria-label="Plan in time order">
          {items.map((item) => (
            <Item key={item.key} item={item} />
          ))}
        </ol>
      )}
    </section>
  );
}
