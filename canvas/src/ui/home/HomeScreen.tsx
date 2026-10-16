import { Boxes, FileJson } from "lucide-react";
import type { PlantMaker } from "../../app/plantMaker";
import type { NavigationStore } from "../../app/navigationStore";
import type { WorkspaceState, WorkspaceStore } from "../../app/workspaceStore";
import type { Workspace } from "../../model/bridge";
import type { Badge } from "../../model/report";
import { workspaceEmpty, workspaceSummary } from "../../model/workspace";
import { Breadcrumb } from "../shell/Breadcrumb";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import styles from "./HomeScreen.module.css";
import { BlankPlantStart } from "./EmptyStart";
import { ExampleList, useExamples } from "./Examples";
import { PlantList } from "./PlantList";
import { RunList } from "./RunList";

interface ContentsProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly badgeFor: (worldId: string) => Badge | null;
}

function EmptyWorkspace({
  name,
  maker,
  navigation,
}: { readonly name: string } & Omit<ContentsProps, "workspace" | "badgeFor">) {
  const examples = useExamples(maker);
  return (
    <div className={styles.empty}>
      <Boxes size={28} aria-hidden="true" />
      <h2 className={styles.emptyTitle}>{`The ${name} workspace has no plants yet`}</h2>
      <p className={styles.lead}>
        Build a simulated plant, generate its readings, test how a connected system responds, and review what happened.
        Plants you create are saved in the {name} workspace folder.
      </p>
      <BlankPlantStart examples={examples} maker={maker} navigation={navigation} />
      <ExampleList examples={examples} taken={[]} maker={maker} navigation={navigation} />
      <ol className={styles.steps}>
        <li>
          Create a new plant{examples.length > 0 ? " or try an example" : ""}, then add its equipment in the editor.
        </li>
        <li>Open the plant and Run it with the faults you want to test.</li>
        <li>Review the run: play it back, read the report and the verdict.</li>
      </ol>
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

function Contents({ workspace, navigation, maker, badgeFor }: ContentsProps) {
  const examples = useExamples(maker);
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <span className={styles.eyebrow}>Workspace folder</span>
        <h2 className={styles.title}>{workspace.name}</h2>
        <span className={styles.summary}>{workspaceSummary(workspace)}</span>
      </header>
      <div className={styles.grid}>
        <PlantList workspace={workspace} navigation={navigation} maker={maker} examples={examples} />
        <div className={styles.column}>
          <RunList workspace={workspace} navigation={navigation} badgeFor={badgeFor} />
          <Scenarios workspace={workspace} />
        </div>
      </div>
    </div>
  );
}

function HomeCrumb({ name }: { readonly name: string }) {
  const heading = useScreen(name === "" ? "Workspace" : `${name} · workspace`);
  return <Breadcrumb trail={[]} here="Workspace" heading={heading} />;
}

function HomeBody({
  state,
  navigation,
  maker,
  badgeFor,
}: { readonly state: WorkspaceState } & Omit<ContentsProps, "workspace">) {
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
        The workspace could not be listed: {state.failure.diagnostic.message} ({state.failure.diagnostic.displayCode})
      </p>
    );
  }
  if (workspaceEmpty(state.workspace)) {
    return <EmptyWorkspace name={state.workspace.name} maker={maker} navigation={navigation} />;
  }
  return <Contents workspace={state.workspace} navigation={navigation} maker={maker} badgeFor={badgeFor} />;
}

export function HomeScreen({
  workspace,
  navigation,
  maker,
  badgeFor = () => null,
}: {
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly badgeFor?: (worldId: string) => Badge | null;
}) {
  const state = useStore(workspace);
  return (
    <>
      <HomeCrumb name={state.phase === "ready" ? state.workspace.name : ""} />
      <HomeBody state={state} navigation={navigation} maker={maker} badgeFor={badgeFor} />
    </>
  );
}
