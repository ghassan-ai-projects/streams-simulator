import { Boxes, FileJson, Radio, Square } from "lucide-react";
import type { ReactNode } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import type { Definition, RunEntry, Workspace } from "../../model/bridge";
import type { Badge } from "../../model/report";
import { badgeLabels } from "../../model/report";
import { workspaceEmpty, workspaceSummary } from "../../model/workspace";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import styles from "./HomeScreen.module.css";

function Section({
  title,
  folder,
  children,
}: {
  readonly title: string;
  readonly folder: string;
  readonly children: ReactNode;
}) {
  return (
    <section className={styles.section} aria-label={title}>
      <h2 className={styles.heading}>
        {title}
        <span className={styles.folder}>{folder}</span>
      </h2>
      {children}
    </section>
  );
}

interface RowsProps {
  readonly items: readonly Definition[];
  readonly empty: string;
  readonly onOpen?: (item: Definition) => void;
}

function DefinitionRows({ items, empty, onOpen }: RowsProps) {
  if (items.length === 0) {
    return <p className={styles.none}>{empty}</p>;
  }
  return (
    <ul className={styles.rows}>
      {items.map((item) => (
        <li key={item.path} className={styles.row}>
          <FileJson size={14} aria-hidden="true" className={styles.icon} />
          <span className={styles.id}>{item.id || item.path}</span>
          <span className={styles.title}>{item.title}</span>
          <span className={styles.path}>{item.path}</span>
          {onOpen && (
            <button
              type="button"
              className={styles.open}
              onClick={() => {
                onOpen(item);
              }}
              aria-label={`Open plant ${item.id || item.path}`}
            >
              Open
            </button>
          )}
        </li>
      ))}
    </ul>
  );
}

interface RunRowsProps {
  readonly runs: readonly RunEntry[];
  readonly onOpen: (run: RunEntry) => void;
  readonly badgeFor: (worldId: string) => Badge | null;
}

function RunBadge({ badge }: { readonly badge: Badge | null }) {
  return (
    <span className={badge === "verified" ? styles.verified : styles.state}>
      {badge === null ? "not checked yet" : badgeLabels[badge]}
    </span>
  );
}

function RunRows({ runs, onOpen, badgeFor }: RunRowsProps) {
  if (runs.length === 0) {
    return <p className={styles.none}>No runs yet. Runs the director writes appear here.</p>;
  }
  return (
    <ul className={styles.rows}>
      {runs.map((run) => (
        <li key={run.worldId} className={styles.row}>
          <span className={styles.id}>{run.worldId}</span>
          <span className={run.explore ? styles.explore : styles.badge}>{run.explore ? "Explore" : "Evaluate"}</span>
          <span className={styles.state}>
            {run.ended ? <Square size={10} aria-hidden="true" /> : <Radio size={12} aria-hidden="true" />}
            {run.ended ? "ended" : "not closed"}
          </span>
          <RunBadge badge={badgeFor(run.worldId)} />
          <button
            type="button"
            className={styles.open}
            onClick={() => {
              onOpen(run);
            }}
            aria-label={`Open review of ${run.worldId}`}
          >
            Review
          </button>
        </li>
      ))}
    </ul>
  );
}

function EmptyWorkspace({ name }: { readonly name: string }) {
  return (
    <div className={styles.empty}>
      <Boxes size={28} aria-hidden="true" />
      <h1 className={styles.emptyTitle}>{name} is empty</h1>
      <p className={styles.lead}>
        No domains, compositions or runs yet. Copy domain files into <code>domains/</code> and compositions into{" "}
        <code>compositions/</code>; they appear here on their own.
      </p>
    </div>
  );
}

interface ContentsProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly badgeFor: (worldId: string) => Badge | null;
}

function Contents({ workspace, navigation, badgeFor }: ContentsProps) {
  const heading = useScreen(workspace.name);
  const openPlant = (item: Definition) => {
    navigation.go({ screen: "plant", domain: item.path, hall: null });
  };
  if (workspaceEmpty(workspace)) {
    return <EmptyWorkspace name={workspace.name} />;
  }
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.name} ref={heading} tabIndex={-1}>
          {workspace.name}
        </h1>
        <span className={styles.summary}>{workspaceSummary(workspace)}</span>
      </header>
      <div className={styles.grid}>
        <Section title="Plants" folder="domains/">
          <DefinitionRows items={workspace.domains} empty="No domains yet." onOpen={openPlant} />
        </Section>
        <Section title="Runs" folder="runs/">
          <RunRows
            badgeFor={badgeFor}
            runs={workspace.runs}
            onOpen={(run) => {
              navigation.go({ screen: "run", worldId: run.worldId });
            }}
          />
        </Section>
        <Section title="Compositions" folder="compositions/">
          <DefinitionRows items={workspace.compositions} empty="No compositions yet." />
        </Section>
        <Section title="Scenarios" folder="scenarios/">
          <DefinitionRows items={workspace.scenarios} empty="No scenarios yet." />
        </Section>
      </div>
    </div>
  );
}

export function HomeScreen({
  workspace,
  navigation,
  badgeFor = () => null,
}: {
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly badgeFor?: (worldId: string) => Badge | null;
}) {
  const state = useStore(workspace);
  if (state.phase === "loading") {
    return (
      <p className={styles.loading} role="status">
        Loading the workspace…
      </p>
    );
  }
  if (state.phase === "failed") {
    return (
      <p className={styles.failure} role="alert">
        The workspace could not be listed: {state.failure.message} ({state.failure.code})
      </p>
    );
  }
  return <Contents workspace={state.workspace} navigation={navigation} badgeFor={badgeFor} />;
}
