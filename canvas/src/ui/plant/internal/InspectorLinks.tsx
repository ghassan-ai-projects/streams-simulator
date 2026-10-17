import type { Link } from "../../../model/plant";
import { hallOf } from "../../../model/graph";
import { spaced } from "../../../model/shared/words";
import styles from "./Inspector.module.css";

function ordered(linked: readonly Link[], side: "in" | "out"): readonly Link[] {
  return [...linked].sort((a, b) => {
    const aID = side === "in" ? a.from : a.to;
    const bID = side === "in" ? b.from : b.to;
    return (
      (hallOf(aID) ?? "").localeCompare(hallOf(bID) ?? "", undefined, { numeric: true }) ||
      aID.localeCompare(bID, undefined, { numeric: true }) ||
      endpointPort(a, side).localeCompare(endpointPort(b, side)) ||
      ownPort(a, side).localeCompare(ownPort(b, side))
    );
  });
}

function endpointPort(link: Link, side: "in" | "out"): string {
  return (side === "in" ? link.fromPort : link.toPort) ?? "";
}

function ownPort(link: Link, side: "in" | "out"): string {
  return (side === "in" ? link.toPort : link.fromPort) ?? "";
}

function endpoint(id: string, port: string | null): string {
  return port === null ? `${id} · port not recorded` : `${id} ${spaced(port)}`;
}

function LinkRow({
  link,
  side,
  onSelect,
}: {
  readonly link: Link;
  readonly side: "in" | "out";
  readonly onSelect: (id: string) => void;
}) {
  const id = side === "in" ? link.from : link.to;
  const relation = `${endpoint(link.from, link.fromPort)} → ${endpoint(link.to, link.toPort)}`;
  return (
    <li>
      <button
        type="button"
        className={styles.linkButton}
        aria-label={`Select ${side === "out" ? "receiver" : "source"} ${id}; ${relation}`}
        onClick={() => {
          onSelect(id);
        }}
      >
        {relation}
      </button>
    </li>
  );
}

function LinkGroup({
  hall,
  links,
  side,
  onSelect,
}: {
  readonly hall: string;
  readonly links: readonly Link[];
  readonly side: "in" | "out";
  readonly onSelect: (id: string) => void;
}) {
  const rows = (
    <ul className={styles.items}>
      {links.map((link, index) => (
        <LinkRow
          key={`${link.from}:${link.fromPort ?? ""}:${link.to}:${link.toPort ?? ""}:${link.type}:${String(index)}`}
          link={link}
          side={side}
          onSelect={onSelect}
        />
      ))}
    </ul>
  );
  if (hall === "") return rows;
  return (
    <section aria-label={`${hall} linked instances`}>
      <h4 className={styles.linkGroup}>{hall}</h4>
      {rows}
    </section>
  );
}

export function InspectorLinks({
  links,
  side,
  onSelect,
}: {
  readonly links: readonly Link[];
  readonly side: "in" | "out";
  readonly onSelect: (id: string) => void;
}) {
  const linked = ordered(links, side);
  if (linked.length === 0) return <p className={styles.empty}>None</p>;
  const groups = new Map<string, Link[]>();
  linked.forEach((link) => {
    const id = side === "in" ? link.from : link.to;
    const group = hallOf(id) ?? "";
    groups.set(group, [...(groups.get(group) ?? []), link]);
  });
  return (
    <>
      {[...groups].map(([hall, group]) => (
        <LinkGroup key={hall} hall={hall} links={group} side={side} onSelect={onSelect} />
      ))}
    </>
  );
}
