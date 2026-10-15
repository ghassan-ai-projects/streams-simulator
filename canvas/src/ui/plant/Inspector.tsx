import type { Instance, Plant } from "../../model/plant";
import type { ReactNode } from "react";
import type { ComponentSpec } from "../../model/ports";
import { linkLines, linksOf } from "../../model/plant";
import { hallMembers } from "../../model/plantView";
import { shortChannel } from "../../model/playback";
import { channelWords } from "../../model/readout";
import { humanise, kindLabel } from "../../model/report";
import { displayUnit } from "../../model/units";
import { X } from "lucide-react";
import type { RunRow } from "../../model/workspace";
import shared from "../shared.module.css";
import { PlantOverview } from "./PlantOverview";
import styles from "./PlantScreen.module.css";
import { InspectorLinks } from "./InspectorLinks";

interface ListProps {
  readonly title: string;
  readonly items: readonly string[];
  readonly mono?: boolean | undefined;
  readonly show?: ((item: string) => string) | undefined;
  readonly count?: number;
  readonly folded?: boolean;
  readonly rows?: ReactNode;
}

function Items({ items, mono = true, show }: Pick<ListProps, "items" | "mono" | "show">) {
  if (items.length === 0) return <p className={styles.empty}>None</p>;
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

function List({ title, items, mono, show, count = items.length, folded = false, rows }: ListProps) {
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
          {rows ?? <Items items={items} mono={mono} show={show} />}
        </details>
      </section>
    );
  }
  return (
    <section className={styles.section} aria-label={title}>
      <h3 className={shared.sectionTitle}>{heading}</h3>
      {rows ?? <Items items={items} mono={mono} show={show} />}
    </section>
  );
}

function channelLine(name: string, channels: readonly { readonly name: string; readonly unit: string }[]): string {
  const unit = displayUnit(channels.find((channel) => channel.name === name)?.unit ?? "");
  return `${channelWords(shortChannel(name))}${unit === "" ? "" : ` · ${unit}`}`;
}

function LinkList({
  title,
  links,
  instance,
  side,
  onSelect,
}: {
  readonly title: string;
  readonly links: ReturnType<typeof linksOf>["incoming"];
  readonly instance: Instance;
  readonly side: "in" | "out";
  readonly onSelect?: ((id: string) => void) | undefined;
}) {
  if (onSelect === undefined)
    return <List title={title} items={linkLines(links, instance.id, side)} count={links.length} mono={false} />;
  return (
    <List
      title={title}
      items={[]}
      count={links.length}
      mono={false}
      rows={<InspectorLinks links={links} side={side} onSelect={onSelect} />}
    />
  );
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

function Head({
  instance,
  spec,
  onClose,
}: {
  readonly instance: Instance;
  readonly spec: ComponentSpec | undefined;
  readonly onClose?: (() => void) | undefined;
}) {
  const title = spec?.title ?? kindLabel(instance.kind);
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
      <p className={styles.kindLine}>{[title, instance.area].filter(Boolean).join(" · ")}</p>
      {instance.component !== null && <p className={styles.source}>{instance.component}</p>}
      {spec?.description !== undefined && (
        <details className={styles.modelDetails}>
          <summary>What this component models</summary>
          <p>{spec.description}</p>
        </details>
      )}
    </header>
  );
}

function InstanceDetails({
  plant,
  instance,
  onClose,
  onSelect,
  spec,
}: {
  readonly plant: Plant;
  readonly instance: Instance;
  readonly onClose?: (() => void) | undefined;
  readonly onSelect?: ((id: string) => void) | undefined;
  readonly spec: ComponentSpec | undefined;
}) {
  const kind = plant.kinds.get(instance.kind);
  const { incoming, outgoing } = linksOf(plant, instance.id);
  return (
    <aside className={styles.inspector} aria-label="Inspector" tabIndex={0}>
      <Head instance={instance} spec={spec} onClose={onClose} />
      <List
        title="Channels"
        items={(kind?.channels ?? []).map((channel) => channel.name)}
        mono={false}
        show={(name) => channelLine(name, kind?.channels ?? [])}
      />
      <LinkList title="Links in" links={incoming} instance={instance} side="in" onSelect={onSelect} />
      <LinkList title="Links out" links={outgoing} instance={instance} side="out" onSelect={onSelect} />
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
  readonly onSelect?: (id: string) => void;
  readonly specs?: ReadonlyMap<string, ComponentSpec>;
}

export function Inspector({ plant, selected, onClose, runs = [], onReview, onSelect, specs }: InspectorProps) {
  const instance = plant.instances.find((entry) => entry.id === selected);
  if (instance !== undefined) {
    const spec = instance.component === null ? undefined : specs?.get(instance.component);
    return <InstanceDetails plant={plant} instance={instance} onClose={onClose} onSelect={onSelect} spec={spec} />;
  }
  const members = selected === null ? [] : hallMembers(plant, selected);
  if (selected !== null && members.length > 0) return <HallSummary hall={selected} members={members} />;
  return <PlantOverview plant={plant} runs={runs} onReview={onReview} />;
}
