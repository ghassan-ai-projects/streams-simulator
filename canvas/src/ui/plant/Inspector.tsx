import type { Instance, Plant } from "../../model/plant";
import { linkLines, linksOf } from "../../model/plant";
import { hallMembers } from "../../model/plantView";
import { shortChannel } from "../../model/playback";
import { channelWords } from "../../model/readout";
import { humanise } from "../../model/report";
import { displayUnit } from "../../model/units";
import { X } from "lucide-react";
import type { RunRow } from "../../model/workspace";
import shared from "../shared.module.css";
import { PlantOverview } from "./PlantOverview";
import styles from "./PlantScreen.module.css";

interface ListProps {
  readonly title: string;
  readonly items: readonly string[];
  readonly mono?: boolean | undefined;
  readonly show?: ((item: string) => string) | undefined;
  readonly count?: number;
  readonly folded?: boolean;
}

function Items({ items, mono = true, show }: Pick<ListProps, "items" | "mono" | "show">) {
  if (items.length === 0) return <p className={styles.muted}>None</p>;
  return (
    <ul className={styles.items}>
      {items.map((item) => (
        <li key={item} className={mono ? styles.mono : undefined} title={show && item}>
          {show ? show(item) : item}
        </li>
      ))}
    </ul>
  );
}

function List({ title, items, mono, show, count = items.length, folded = false }: ListProps) {
  const heading = (
    <>
      {title}
      <span className={styles.count}>{count}</span>
    </>
  );
  if (folded && items.length > 0) {
    return (
      <section className={styles.section} aria-label={title}>
        <details>
          <summary className={shared.sectionTitle}>{heading}</summary>
          <Items items={items} mono={mono} show={show} />
        </details>
      </section>
    );
  }
  return (
    <section className={styles.section} aria-label={title}>
      <h3 className={shared.sectionTitle}>{heading}</h3>
      <Items items={items} mono={mono} show={show} />
    </section>
  );
}

function channelLine(name: string, channels: readonly { readonly name: string; readonly unit: string }[]): string {
  const unit = displayUnit(channels.find((channel) => channel.name === name)?.unit ?? "");
  return `${channelWords(shortChannel(name))}${unit === "" ? "" : ` · ${unit}`}`;
}

function HallSummary({
  hall,
  members,
}: {
  readonly hall: string;
  readonly members: readonly (readonly [string, number])[];
}) {
  const total = members.reduce((sum, [, n]) => sum + n, 0);
  return (
    <aside className={styles.inspector} aria-label="Inspector" tabIndex={0}>
      <header className={styles.inspectorHead}>
        <h2 className={styles.mono}>{hall}</h2>
        <p className={styles.muted}>{`Hall · ${String(total)} instances · Enter or double-click opens it`}</p>
      </header>
      <List
        title="Members"
        items={members.map(([kind, n]) => `${channelWords(kind)} × ${String(n)}`)}
        mono={false}
        count={total}
      />
    </aside>
  );
}

function Head({ instance, onClose }: { readonly instance: Instance; readonly onClose?: (() => void) | undefined }) {
  return (
    <header className={styles.inspectorHead}>
      <div className={styles.headLine}>
        <h2 className={styles.mono}>{instance.id}</h2>
        {onClose !== undefined && (
          <button type="button" className={styles.close} aria-label="Close the inspector" onClick={onClose}>
            <X size={16} aria-hidden="true" />
          </button>
        )}
      </div>
      <p className={styles.kindLine}>{[humanise(instance.kind), instance.area].filter(Boolean).join(" · ")}</p>
      {instance.component !== null && <p className={styles.source}>{instance.component}</p>}
    </header>
  );
}

function InstanceDetails({
  plant,
  instance,
  onClose,
}: {
  readonly plant: Plant;
  readonly instance: Instance;
  readonly onClose?: (() => void) | undefined;
}) {
  const kind = plant.kinds.get(instance.kind);
  const { incoming, outgoing } = linksOf(plant, instance.id);
  return (
    <aside className={styles.inspector} aria-label="Inspector" tabIndex={0}>
      <Head instance={instance} onClose={onClose} />
      <List
        title="Channels"
        items={(kind?.channels ?? []).map((channel) => channel.name)}
        mono={false}
        show={(name) => channelLine(name, kind?.channels ?? [])}
      />
      <List title="Links in" items={linkLines(incoming, instance.id, "in")} count={incoming.length} mono={false} />
      <List title="Links out" items={linkLines(outgoing, instance.id, "out")} count={outgoing.length} mono={false} />
      <List title="Faults" items={kind?.faults ?? []} mono={false} show={humanise} folded />
      <List title="Effectors" items={kind?.effectors ?? []} mono={false} show={humanise} />
      <List title="States" items={kind?.states ?? []} mono={false} show={channelWords} folded />
    </aside>
  );
}

interface InspectorProps {
  readonly plant: Plant;
  readonly selected: string | null;
  readonly onClose?: () => void;
  readonly runs?: readonly RunRow[];
  readonly onReview?: (worldId: string) => void;
}

export function Inspector({ plant, selected, onClose, runs = [], onReview }: InspectorProps) {
  const instance = plant.instances.find((entry) => entry.id === selected);
  if (instance !== undefined) return <InstanceDetails plant={plant} instance={instance} onClose={onClose} />;
  const members = selected === null ? [] : hallMembers(plant, selected);
  if (selected !== null && members.length > 0) return <HallSummary hall={selected} members={members} />;
  return <PlantOverview plant={plant} runs={runs} onReview={onReview} />;
}
