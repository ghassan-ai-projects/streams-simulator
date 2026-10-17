import { shared } from "../../primitives";
import type { Instance, Plant } from "../../../model/plant";
import type { ComponentSpec } from "../../../model/plant";
import { linksOf } from "../../../model/plant";
import { channelWords, hallMembers } from "../../../model/graph";
import { shortChannel } from "../../../model/shared/words";
import { humanise, kindLabel } from "../../../model/shared/words";
import { displayUnit } from "../../../model/shared/units";
import { X } from "lucide-react";
import type { RunRow } from "../../../model/workspace";
import { cx } from "../../primitives";
import { useCutBottom } from "../../primitives";
import { PlantOverview } from "./PlantOverview";
import styles from "./Inspector.module.css";
import { InspectorLinks } from "./InspectorLinks";
import { List } from "./InspectorList";

function channelLine(name: string, channels: readonly { readonly name: string; readonly unit: string }[]): string {
  const unit = displayUnit(channels.find((channel) => channel.name === name)?.unit ?? "");
  return `${channelWords(shortChannel(name))}${unit === "" ? "" : ` · ${unit}`}`;
}

function LinkList({
  title,
  links,
  side,
  onSelect,
}: {
  readonly title: string;
  readonly links: ReturnType<typeof linksOf>["incoming"];
  readonly side: "in" | "out";
  readonly onSelect: (id: string) => void;
}) {
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
  const { ref, cut } = useCutBottom<HTMLElement>();
  return (
    <aside
      ref={ref}
      className={cx(styles.inspector, shared.scrolls, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-label="Inspector"
      tabIndex={0}
    >
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
  readonly onClose: () => void;
}) {
  const title = spec?.title ?? kindLabel(instance.kind);
  return (
    <header className={styles.inspectorHead}>
      <div className={styles.headLine}>
        <h2 className={styles.mono}>{instance.id}</h2>
        <button type="button" className={styles.close} aria-label="Close the inspector" onClick={onClose}>
          <X size={16} aria-hidden="true" />
        </button>
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
  readonly onClose: () => void;
  readonly onSelect: (id: string) => void;
  readonly spec: ComponentSpec | undefined;
}) {
  const kind = plant.kinds.get(instance.kind);
  const { incoming, outgoing } = linksOf(plant, instance.id);
  const { ref, cut } = useCutBottom<HTMLElement>();
  return (
    <aside
      ref={ref}
      className={cx(styles.inspector, shared.scrolls, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-label="Inspector"
      tabIndex={0}
    >
      <Head instance={instance} spec={spec} onClose={onClose} />
      <List
        title="Channels"
        items={(kind?.channels ?? []).map((channel) => channel.name)}
        mono={false}
        show={(name) => channelLine(name, kind?.channels ?? [])}
      />
      <LinkList title="Links in" links={incoming} side="in" onSelect={onSelect} />
      <LinkList title="Links out" links={outgoing} side="out" onSelect={onSelect} />
      <List title="Faults" items={kind?.faults ?? []} mono={false} show={humanise} folded />
      <List title="Effectors" items={kind?.effectors ?? []} mono={false} show={humanise} />
      <List title="States" items={kind?.states ?? []} mono={false} show={channelWords} folded />
    </aside>
  );
}

interface InspectorProps {
  readonly plant: Plant;
  readonly selected: string | null;
  readonly onClose: () => void;
  readonly runs: readonly RunRow[];
  readonly onReview: (worldId: string) => void;
  readonly onSelect: (id: string) => void;
  readonly specs: ReadonlyMap<string, ComponentSpec>;
}

export function Inspector({ plant, selected, onClose, runs, onReview, onSelect, specs }: InspectorProps) {
  const instance = plant.instances.find((entry) => entry.id === selected);
  if (instance !== undefined) {
    const spec = instance.component === null ? undefined : specs.get(instance.component);
    return <InstanceDetails plant={plant} instance={instance} onClose={onClose} onSelect={onSelect} spec={spec} />;
  }
  const members = selected === null ? [] : hallMembers(plant, selected);
  if (selected !== null && members.length > 0) return <HallSummary hall={selected} members={members} />;
  return <PlantOverview plant={plant} runs={runs} onReview={onReview} />;
}
