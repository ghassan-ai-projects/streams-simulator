import type { Loaded } from "../../../app/launch";
import type { DeliveryProblem, Draft, DraftPerturbation, PerturbationKind, PlanProblem } from "../../../model/launch";
import { kindNamed } from "../../../model/launch";
import { humanise } from "../../../model/shared/words";
import { panel, useFocusOnMount } from "../../primitives";
import { DeliveryParams } from "./DeliveryParams";
import type { Change } from "./RowParts";
import { AddRow, invalid, ProblemNote, RowTail, RowTimes, useAdding } from "./RowParts";
import styles from "./LaunchPanel.module.css";
import rows from "./PlanRows.module.css";

type Catalog = Loaded<readonly PerturbationKind[]>;

interface KindPickProps {
  readonly delivery: DraftPerturbation;
  readonly kinds: readonly PerturbationKind[];
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly fresh: boolean;
  readonly onPick: (kind: string) => void;
}

function KindPick({ delivery, kinds, label, id, problem, fresh, onPick }: KindPickProps) {
  const first = useFocusOnMount<HTMLSelectElement>({ when: fresh });
  const stale = delivery.kind !== "" && kindNamed(kinds, delivery.kind) === undefined;
  return (
    <select
      ref={first}
      id={id}
      className={styles.instance}
      aria-label={`${label} kind`}
      value={delivery.kind}
      {...invalid(problem, id)}
      onChange={(event) => {
        onPick(event.target.value);
      }}
    >
      <option value="">Choose a delivery condition</option>
      {stale && <option value={delivery.kind}>{humanise(delivery.kind)}</option>}
      {kinds.map((kind) => (
        <option key={kind.kind} value={kind.kind}>
          {humanise(kind.kind)}
        </option>
      ))}
    </select>
  );
}

function KindNote({ kind }: { readonly kind: PerturbationKind | undefined }) {
  if (kind === undefined) return null;
  const end = kind.requiresEnd ? " It needs an end minute." : "";
  return (
    <p className={rows.note}>
      <span className={rows.kind}>Delivery</span> {kind.description}
      {end}
    </p>
  );
}

interface RowProps {
  readonly delivery: DraftPerturbation;
  readonly kinds: readonly PerturbationKind[];
  readonly index: number;
  readonly problem: DeliveryProblem | null;
  readonly change: (change: Change) => void;
  readonly fresh: boolean;
  readonly onRemoved: () => void;
}

function useDeliveryEdit(change: (change: Change) => void, key: number, onRemoved: () => void) {
  const update = (next: Partial<DraftPerturbation>) => {
    change((d) => ({ ...d, perturbations: d.perturbations.map((p) => (p.key === key ? { ...p, ...next } : p)) }));
  };
  const remove = () => {
    change((d) => ({ ...d, perturbations: d.perturbations.filter((p) => p.key !== key) }));
    onRemoved();
  };
  return { update, remove };
}

function DeliveryRow({ delivery, kinds, index, problem, change, fresh, onRemoved }: RowProps) {
  const label = `Delivery condition ${String(index + 1)}`;
  const id = `launch-delivery-${String(index + 1)}`;
  const { update, remove } = useDeliveryEdit(change, delivery.key, onRemoved);
  const kind = kindNamed(kinds, delivery.kind);
  const part = (name: DeliveryProblem["part"]) => (problem?.part === name ? problem.text : "");
  return (
    <li className={styles.faultRow} aria-label={label}>
      <RowTimes
        label={label}
        id={id}
        atMinutes={delivery.atMinutes}
        untilMinutes={delivery.untilMinutes}
        startProblem={part("start")}
        endProblem={part("end")}
        endPlaceholder={kind?.requiresEnd === true ? "needed" : "run end"}
        update={update}
      />
      <KindPick
        delivery={delivery}
        kinds={kinds}
        label={label}
        id={id}
        problem={part("kind")}
        fresh={fresh}
        onPick={(picked) => {
          update({ kind: picked, values: {} });
        }}
      />
      <KindNote kind={kind} />
      <DeliveryParams delivery={delivery} kind={kind} label={label} id={id} problem={part("param")} update={update} />
      <RowTail label={label} id={id} problem={problem?.part === "end" ? "" : (problem?.text ?? "")} onRemove={remove} />
      <ProblemNote id={`${id}-end`} text={part("end")} />
    </li>
  );
}

const problemOf = (problems: readonly PlanProblem[], index: number): DeliveryProblem | null =>
  problems.find((p): p is DeliveryProblem => p.field === "delivery" && p.index === index) ?? null;

function CatalogNote({ catalog }: { readonly catalog: Catalog }) {
  if (catalog.status === "ready") return null;
  const words =
    catalog.status === "loading"
      ? "Reading the delivery conditions the simulator offers…"
      : `The delivery conditions could not be read: ${catalog.words}`;
  return (
    <p className={panel.hint} role={catalog.status === "unavailable" ? "alert" : undefined}>
      {words}
    </p>
  );
}

interface RowsProps {
  readonly draft: Draft;
  readonly catalog: Catalog;
  readonly problems: readonly PlanProblem[];
  readonly change: (change: Change) => void;
}

const appended = (d: Draft, key: number): Draft => ({
  ...d,
  perturbations: [...d.perturbations, { key, kind: "", atMinutes: "0", untilMinutes: "", values: {} }],
});

export function DeliveryRows({ draft, catalog, problems, change }: RowsProps) {
  const { added, addButton, add } = useAdding(change, appended);
  const kinds = catalog.status === "ready" ? catalog.value : [];
  return (
    <fieldset className={panel.group}>
      <legend>Delivery conditions</legend>
      <p className={panel.hint}>
        What the outside system receives, not what happens in the plant: records lost, late, repeated or out of order.
      </p>
      <CatalogNote catalog={catalog} />
      <ol className={styles.faults}>
        {draft.perturbations.map((delivery, index) => (
          <DeliveryRow
            key={delivery.key}
            delivery={delivery}
            kinds={kinds}
            index={index}
            problem={problemOf(problems, index)}
            change={change}
            fresh={delivery.key === added}
            onRemoved={() => {
              addButton.current?.focus();
            }}
          />
        ))}
      </ol>
      <AddRow
        label="Add delivery condition"
        buttonRef={addButton}
        disabled={catalog.status !== "ready"}
        onAdd={() => {
          add(draft.nextKey);
        }}
      />
    </fieldset>
  );
}
