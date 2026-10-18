import { shared } from "../../primitives";
import type { OutlineProps } from "./types";
import { OctagonAlert, Search } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { groups, matchWords, marked, belowCount, hallsOf } from "../../../model/graph";
import type { InstanceOrigin, OutlineItem as Item } from "../../../model/graph";
import { OriginMark } from "./OriginMark";
import { cx } from "../../primitives";
import { useCutBottom } from "../../primitives";
import styles from "./graph.module.css";

interface SectionProps {
  readonly title: string;
  readonly depth?: number;
  readonly items: readonly Item[];
  readonly below?: number;
  readonly current: string | null;
  readonly marks: "location" | "selection";
  readonly onPick: (item: string) => void;
  readonly origins?: ReadonlyMap<string, InstanceOrigin> | undefined;
}

function countWords(direct: number, below: number): string {
  if (direct > 0) return String(direct);
  return below > 0 ? `0 · ${String(below)} below` : "0";
}

function countTitle(direct: number, below: number): string | undefined {
  if (direct > 0) return undefined;
  return below > 0 ? `${String(below)} in child areas` : "No equipment yet";
}

function Entry({
  item,
  current,
  marks,
  onPick,
  origins,
}: Omit<SectionProps, "title" | "items"> & { readonly item: Item }) {
  const button = useRef<HTMLButtonElement>(null);
  const isCurrent = item.id === current;
  useEffect(() => {
    if (isCurrent) button.current?.scrollIntoView({ block: "nearest" });
  }, [isCurrent]);
  return (
    <li>
      <button
        ref={button}
        type="button"
        className={isCurrent ? styles.current : styles.item}
        aria-current={isCurrent ? (marks === "location" ? "location" : "true") : undefined}
        onClick={() => {
          onPick(item.id);
        }}
      >
        <span className={styles.itemId}>{item.id}</span>
        {item.kind !== "" && <span className={styles.itemKind}>{item.kind}</span>}
        <OriginMark origin={origins?.get(item.id) ?? null} variant="glyph" />
        {item.flagged === true && (
          <span className={styles.itemIssue}>
            <OctagonAlert size={12} aria-hidden="true" />
            <span className="sr-only">has issues</span>
          </span>
        )}
      </button>
    </li>
  );
}

function Section({ title, depth = 0, items, below = 0, current, marks, onPick, origins }: SectionProps) {
  return (
    <details open style={{ marginLeft: `calc(var(--space-3) * ${String(depth)})` }}>
      <summary className={cx(styles.areaRow, shared.disclosure)}>
        <span className={styles.areaName} title={title}>
          {title}
        </span>
        <span className={styles.count} title={countTitle(items.length, below)}>
          {countWords(items.length, below)}
        </span>
      </summary>
      {items.length > 0 && (
        <ul className={styles.tree}>
          {items.map((item) => (
            <Entry key={item.id} item={item} current={current} marks={marks} onPick={onPick} origins={origins} />
          ))}
        </ul>
      )}
    </details>
  );
}

function Filter({ value, onChange }: { readonly value: string; readonly onChange: (value: string) => void }) {
  return (
    <label className={styles.filter}>
      <Search size={14} aria-hidden="true" />
      <input
        type="search"
        placeholder="Filter by id or kind"
        aria-label="Filter instances"
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    </label>
  );
}

interface AreaSectionsProps {
  readonly shown: ReturnType<typeof marked>;
  readonly selected: string | null;
  readonly onSelect: (id: string) => void;
  readonly origins: ReadonlyMap<string, InstanceOrigin> | undefined;
}

function AreaSections({ shown, selected, onSelect, origins }: AreaSectionsProps) {
  return (
    <>
      {shown.map(({ area, depth, items }, index) => (
        <Section
          key={area}
          title={area}
          depth={depth}
          items={items}
          below={belowCount(shown, index)}
          current={selected}
          marks="selection"
          onPick={onSelect}
          origins={origins}
        />
      ))}
    </>
  );
}

export function Outline({
  plant,
  hall,
  selected,
  onSelect,
  onOpenHall,
  flagged,
  showEmptyAreas,
  origins,
}: OutlineProps) {
  const [filter, setFilter] = useState("");
  const { ref: navRef, cut } = useCutBottom<HTMLElement>();
  const halls = hallsOf(plant).map((id) => ({ id, kind: "" }));
  const shown = marked(groups(plant, hall, filter, { emptyAreas: showEmptyAreas === true ? "show" : "hide" }), flagged);
  return (
    <nav
      ref={navRef}
      className={cx(styles.outline, shared.cut)}
      aria-label="Outline"
      data-cut={cut ? "true" : undefined}
    >
      <Filter value={filter} onChange={setFilter} />
      {halls.length > 0 && onOpenHall !== undefined && filter === "" && (
        <Section title="Halls" items={halls} current={hall} marks="location" onPick={onOpenHall} />
      )}
      <AreaSections shown={shown} selected={selected} onSelect={onSelect} origins={origins} />
      <p className={shown.length === 0 ? styles.noMatch : "sr-only"} role="status">
        {matchWords(shown, filter, plant.instances.length === 0)}
      </p>
    </nav>
  );
}
