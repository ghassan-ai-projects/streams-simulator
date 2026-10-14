import type { Link } from "../../model/plant";
import { hallOf } from "../../model/plantView";
import styles from "./PlantScreen.module.css";

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
  return (
    <section aria-label={`${hall} linked instances`}>
      <h4 className={styles.linkGroup}>{hall}</h4>
      <ul className={styles.items}>
        {links.map((link, index) => {
          const id = side === "in" ? link.from : link.to;
          const relation = `${endpoint(link.from, link.fromPort)} → ${endpoint(link.to, link.toPort)}`;
          return (
            <li
              key={`${link.from}:${link.fromPort ?? ""}:${link.to}:${link.toPort ?? ""}:${link.type}:${String(index)}`}
            >
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
        })}
      </ul>
    </section>
  );
}

function endpoint(id: string, port: string | null): string {
  return port === null ? `${id} · port not recorded` : `${id} ${port.replaceAll("_", " ")}`;
}

export function InspectorLinks({
  links,
  side,
  onSelect,
}: {
  readonly links: readonly Link[];
  readonly side: "in" | "out";
  readonly onSelect?: ((id: string) => void) | undefined;
}) {
  const linked = ordered(links, side);
  if (linked.length === 0 || onSelect === undefined) return null;
  const groups = new Map<string, Link[]>();
  linked.forEach((link) => {
    const id = side === "in" ? link.from : link.to;
    const group = hallOf(id) ?? "Site";
    groups.set(group, [...(groups.get(group) ?? []), link]);
  });
  return (
    <details className={styles.linkDetails}>
      <summary className={styles.linkSummary}>
        Browse {side === "out" ? "outgoing links" : "incoming links"} ({String(linked.length)})
      </summary>
      {[...groups].map(([hall, group]) => (
        <LinkGroup key={hall} hall={hall} links={group} side={side} onSelect={onSelect} />
      ))}
    </details>
  );
}
