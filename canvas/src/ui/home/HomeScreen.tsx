import { Boxes, FileJson } from "lucide-react";
import type { EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import type { Workspace } from "../../model/bridge";
import type { Badge } from "../../model/report";
import { workspaceEmpty, workspaceSummary } from "../../model/workspace";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import styles from "./HomeScreen.module.css";
import { PlantList } from "./PlantList";
import { RunList } from "./RunList";

interface ContentsProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly editor: EditorStore;
  readonly badgeFor: (worldId: string) => Badge | null;
}

function EmptyWorkspace({ name }: { readonly name: string }) {
  const heading = useScreen(name);
  return (
    <div className={styles.empty}>
      <Boxes size={28} aria-hidden="true" />
      <h1 className={styles.emptyTitle} ref={heading} tabIndex={-1}>
        {`${name} has no plants yet`}
      </h1>
      <ol className={styles.steps}>
        <li>Copy a plant into the workspace: a composition into compositions/, or a domain into domains/.</li>
        <li>Open it here, then Run it with the faults you want to test.</li>
        <li>Review the run: play it back, read the report and the verdict.</li>
      </ol>
      <p className={styles.lead}>Files you add appear here on their own.</p>
    </div>
  );
}

function Scenarios({ workspace }: { readonly workspace: Workspace }) {
  if (workspace.scenarios.length === 0) return null;
  return (
    <section className={styles.section} aria-labelledby="scenarios-title">
      <header className={styles.heading}>
        <h2 id="scenarios-title">Scenarios</h2>
        <span className={styles.count}>{workspace.scenarios.length}</span>
      </header>
      <ul className={styles.rows}>
        {workspace.scenarios.map((item) => (
          <li key={item.path} className={styles.item} title={item.path}>
            <FileJson size={16} aria-hidden="true" className={styles.icon} />
            <span className={styles.name}>{item.title || item.id || item.path}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Contents({ workspace, navigation, editor, badgeFor }: ContentsProps) {
  const heading = useScreen(workspace.name);
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title} ref={heading} tabIndex={-1}>
          {workspace.name}
        </h1>
        <span className={styles.summary}>{workspaceSummary(workspace)}</span>
      </header>
      <div className={styles.grid}>
        <PlantList workspace={workspace} navigation={navigation} editor={editor} />
        <div className={styles.column}>
          <RunList workspace={workspace} navigation={navigation} badgeFor={badgeFor} />
          <Scenarios workspace={workspace} />
        </div>
      </div>
    </div>
  );
}

export function HomeScreen({
  workspace,
  navigation,
  editor,
  badgeFor = () => null,
}: {
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly editor: EditorStore;
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
  if (workspaceEmpty(state.workspace)) return <EmptyWorkspace name={state.workspace.name} />;
  return <Contents workspace={state.workspace} navigation={navigation} editor={editor} badgeFor={badgeFor} />;
}
