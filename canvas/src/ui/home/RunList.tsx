import { Eye, Lock, Radio, TriangleAlert } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { Workspace } from "../../model/bridge";
import type { Badge } from "../../model/report";
import { badgeLabels } from "../../model/report";
import type { RunRow } from "../../model/workspace";
import { plantsOf, runFacts, runRows, unfinishedReason, whenOf } from "../../model/workspace";
import styles from "./HomeScreen.module.css";
import { MetaParts } from "../MetaParts";

interface ListProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly badgeFor: (worldId: string) => Badge | null;
}

const page = 20;

function moreLabel(count: number): string {
  return `Show ${String(count)} more ${count === 1 ? "run" : "runs"}`;
}

function detailParts(row: RunRow): readonly string[] {
  return [row.run.worldId, whenOf(row.run.recordedAt), ...runFacts(row.run)].filter((part) => part !== "");
}

function detail(row: RunRow): string {
  return detailParts(row).join(" · ");
}

function Status({ row, badge }: { readonly row: RunRow; readonly badge: Badge | null }) {
  if (!row.run.ended) {
    return (
      <span className={styles.live}>
        <Radio size={12} aria-hidden="true" />
        not finished
      </span>
    );
  }
  if (!row.run.complete) {
    return (
      <span className={styles.badge} title={unfinishedReason(row.run)}>
        <TriangleAlert size={12} aria-hidden="true" />
        {badgeLabels.incomplete}
      </span>
    );
  }
  const state = badge ?? "not-checked";
  if (state === "not-checked") return null;
  return <span className={state === "verified" ? styles.verified : styles.badge}>{badgeLabels[state]}</span>;
}

function unchecked(rows: readonly RunRow[], badgeFor: ListProps["badgeFor"]): boolean {
  return rows.some((row) => row.run.complete && (badgeFor(row.run.worldId) ?? "not-checked") === "not-checked");
}

function Row({
  row,
  badge,
  navigation,
}: { readonly row: RunRow; readonly badge: Badge | null } & Pick<ListProps, "navigation">) {
  return (
    <li className={styles.item}>
      <span className={styles.lines}>
        <span className={styles.line}>
          <span className={styles.name} title={row.plantTitle}>
            {row.plantTitle}
          </span>
          <span className={row.run.explore ? styles.explore : styles.badge}>
            {row.run.explore ? <Eye size={12} aria-hidden="true" /> : <Lock size={12} aria-hidden="true" />}
            {row.run.explore ? "Explore" : "Evaluate"}
          </span>
        </span>
        <span className={styles.meta} title={detail(row)}>
          <MetaParts parts={detailParts(row)} />
        </span>
      </span>
      <Status row={row} badge={badge} />
      <span className={styles.actions}>
        <button
          type="button"
          className={styles.open}
          aria-label={`Review ${row.run.worldId}`}
          title="Review checks recording integrity using the run files."
          onClick={() => {
            navigation.go({ screen: "run", worldId: row.run.worldId });
          }}
        >
          Review
        </button>
      </span>
    </li>
  );
}

function PlantFilter({
  workspace,
  plant,
  setPlant,
}: {
  readonly workspace: Workspace;
  readonly plant: string;
  readonly setPlant: (plant: string) => void;
}) {
  const plants = plantsOf(workspace).filter((entry) => entry.runs > 0);
  if (plants.length < 2) return null;
  return (
    <select
      className={styles.filter}
      aria-label="Show runs of"
      value={plant}
      onChange={(event) => {
        setPlant(event.target.value);
      }}
    >
      <option value="">All plants</option>
      {plants.map((entry) => (
        <option key={entry.id} value={entry.id}>
          {entry.title}
        </option>
      ))}
    </select>
  );
}

function useMore() {
  const [shown, setShown] = useState(page);
  const list = useRef<HTMLUListElement>(null);
  const [focusAt, setFocusAt] = useState<number | null>(null);
  useEffect(() => {
    if (focusAt !== null) list.current?.querySelectorAll<HTMLButtonElement>("button")[focusAt]?.focus();
  }, [focusAt]);
  const more = () => {
    setFocusAt(shown);
    setShown(shown + page);
  };
  return { shown, list, more };
}

export function RunList({ workspace, navigation, badgeFor }: ListProps) {
  const [plant, setPlant] = useState("");
  const { shown, list, more } = useMore();
  const rows = runRows(workspace, plant === "" ? null : plant);
  return (
    <section className={styles.section} aria-labelledby="runs-title">
      <header className={styles.heading}>
        <h2 id="runs-title">Runs</h2>
        <span className={styles.count}>{rows.length}</span>
        <PlantFilter workspace={workspace} plant={plant} setPlant={setPlant} />
      </header>
      {unchecked(rows, badgeFor) && <p className={styles.note}>Runs are checked when you open their review.</p>}
      {rows.length === 0 && <p className={styles.none}>No runs yet. Open a plant and choose Run.</p>}
      <p className="sr-only" role="status">
        {plant === "" ? "" : `${String(rows.length)} ${rows.length === 1 ? "run" : "runs"} shown`}
      </p>
      <ul className={styles.rows} ref={list}>
        {rows.slice(0, shown).map((row) => (
          <Row key={row.run.worldId} row={row} badge={badgeFor(row.run.worldId)} navigation={navigation} />
        ))}
      </ul>
      {rows.length > shown && (
        <button type="button" className={styles.more} onClick={more}>
          {moreLabel(Math.min(page, rows.length - shown))}
        </button>
      )}
    </section>
  );
}
