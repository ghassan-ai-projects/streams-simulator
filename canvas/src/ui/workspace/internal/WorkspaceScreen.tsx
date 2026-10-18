import { Boxes, FileJson } from "lucide-react";
import { useState } from "react";
import type { DeviceDesk } from "../../../app/devices";
import type { PlantMaker } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import type { WorkspaceState, WorkspaceStore } from "../../../app/workspace";
import type { Workspace } from "../../../model/workspace";
import type { Badge } from "../../../model/review";
import { workspaceEmpty, workspaceSummary } from "../../../model/workspace";
import { Breadcrumb, failureLine, screen } from "../../primitives";
import { useScreen } from "../../primitives";
import { useStore } from "../../primitives";
import rowStyles from "./ListSection.module.css";
import styles from "./WorkspaceScreen.module.css";
import { CaptureScreen } from "./CaptureScreen";
import { DeviceList } from "./DeviceList";
import type { CaptureEntry } from "./DeviceList";
import { BlankPlantStart } from "./EmptyStart";
import { ExampleList, useExamples } from "./Examples";
import { PlantList } from "./PlantList";
import { RunList } from "./RunList";
import { SectionHeading } from "./SectionHeading";

interface ContentsProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly desk: DeviceDesk;
  readonly capture: CaptureEntry;
  readonly badgeFor: (worldId: string) => Badge | null;
}

function EmptyWorkspace({ workspace, maker, desk, navigation, capture }: Omit<ContentsProps, "badgeFor">) {
  const examples = useExamples(maker);
  const { name } = workspace;
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
      <div className={styles.deviceStrip}>
        <DeviceList workspace={workspace} navigation={navigation} desk={desk} capture={capture} />
      </div>
    </div>
  );
}

function Scenarios({ workspace }: { readonly workspace: Workspace }) {
  if (workspace.scenarios.length === 0) return null;
  return (
    <section className={rowStyles.section} aria-labelledby="scenarios-title">
      <SectionHeading id="scenarios-title" title="Scenarios" count={workspace.scenarios.length} />
      <ul className={rowStyles.rows}>
        {workspace.scenarios.map((item) => (
          <li key={item.path} className={rowStyles.item} title={item.path}>
            <FileJson size={16} aria-hidden="true" className={rowStyles.icon} />
            <span className={rowStyles.name}>{item.title || item.id || item.path}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Contents({ workspace, navigation, maker, desk, capture, badgeFor }: ContentsProps) {
  const examples = useExamples(maker);
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <span className={styles.eyebrow}>Workspace folder</span>
        <h2 className={styles.title}>{workspace.name}</h2>
        <span className={styles.summary}>{workspaceSummary(workspace)}</span>
      </header>
      <div className={styles.grid}>
        <div className={styles.column}>
          <PlantList workspace={workspace} navigation={navigation} maker={maker} examples={examples} />
          <DeviceList workspace={workspace} navigation={navigation} desk={desk} capture={capture} />
        </div>
        <div className={styles.column}>
          <RunList workspace={workspace} navigation={navigation} badgeFor={badgeFor} />
          <Scenarios workspace={workspace} />
        </div>
      </div>
    </div>
  );
}

function WorkspaceCrumb({ name }: { readonly name: string }) {
  const heading = useScreen(name === "" ? "Workspace" : `${name} · workspace`);
  return <Breadcrumb trail={[]} here="Workspace" heading={heading} />;
}

function ReadyBody(props: ContentsProps) {
  return workspaceEmpty(props.workspace) ? <EmptyWorkspace {...props} /> : <Contents {...props} />;
}

function WorkspaceBody({ state, ...rest }: { readonly state: WorkspaceState } & Omit<ContentsProps, "workspace">) {
  if (state.phase === "loading") {
    return (
      <p className={screen.loading} role="status">
        Loading the workspace…
      </p>
    );
  }
  if (state.phase === "failed") {
    return (
      <p className={screen.failure} role="alert">
        The workspace could not be listed: {failureLine(state.failure.diagnostic)}
      </p>
    );
  }
  return <ReadyBody workspace={state.workspace} {...rest} />;
}

function useCaptureEntry() {
  const [capturing, setCapturing] = useState<"closed" | "open" | "left">("closed");
  const entry: CaptureEntry = {
    onOpen: () => {
      setCapturing("open");
    },
    justLeft: capturing === "left",
  };
  return {
    entry,
    open: capturing === "open",
    leave: () => {
      setCapturing("left");
    },
  };
}

export function WorkspaceScreen({
  workspace,
  navigation,
  maker,
  desk,
  badgeFor,
}: {
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly desk: DeviceDesk;
  readonly badgeFor: (worldId: string) => Badge | null;
}) {
  const state = useStore(workspace);
  const { entry, open, leave } = useCaptureEntry();
  if (open && state.phase === "ready") {
    return <CaptureScreen workspace={state.workspace} desk={desk} navigation={navigation} onClose={leave} />;
  }
  return (
    <>
      <WorkspaceCrumb name={state.phase === "ready" ? state.workspace.name : ""} />
      <WorkspaceBody
        state={state}
        navigation={navigation}
        maker={maker}
        desk={desk}
        capture={entry}
        badgeFor={badgeFor}
      />
    </>
  );
}
