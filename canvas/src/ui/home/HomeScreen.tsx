import { Boxes, FileJson, Play, Radio } from "lucide-react";
import type { ReactNode } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import type { Definition, RunEntry, Workspace } from "../../model/bridge";
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

function RunRows({ runs, onOpen }: { readonly runs: readonly RunEntry[]; readonly onOpen: (run: RunEntry) => void }) {
  if (runs.length === 0) {
    return <p className={styles.none}>No runs yet. Runs the director writes appear here.</p>;
  }
  return (
    <ul className={styles.rows}>
      {runs.map((run) => (
        <li key={run.worldId} className={styles.row}>
          <span className={styles.id}>{run.worldId}</span>
          <span className={styles.badge}>{run.explore ? "Explore" : "Evaluate"}</span>
          <span className={run.ended ? styles.state : styles.live}>
            {run.ended ? <Play size={12} aria-hidden="true" /> : <Radio size={12} aria-hidden="true" />}
            {run.ended ? "ended" : "running"}
          </span>
          {run.ended && (
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
          )}
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

function Contents({ workspace, navigation }: { readonly workspace: Workspace; readonly navigation: NavigationStore }) {
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
}: {
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
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
  return <Contents workspace={state.workspace} navigation={navigation} />;
}
