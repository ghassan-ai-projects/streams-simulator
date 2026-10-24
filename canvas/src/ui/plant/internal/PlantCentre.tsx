import { useMemo } from "react";
import { cx, screen, shared, useStore } from "../../primitives";
import type { ComponentProps } from "react";
import { Network, Radio } from "lucide-react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { PlantSession, ReadyPlant } from "../../../app/plant";
import { hallCount, plantView } from "../../../model/graph";
import { plantTitle } from "../../../model/plant";
import { counted } from "../../../model/workspace";
import { shownFaults } from "../../../model/launch";
import type { Crumb } from "../../primitives";
import { Breadcrumb } from "../../primitives";
import { useScreen } from "../../primitives";
import { ZoomBar } from "../../graph";
import { EditButton, RunButton, useRunningHere } from "./PlantActions";
import { PlantCanvas } from "./PlantCanvas";
import styles from "./PlantBar.module.css";
import type { Connected, PlannedFault } from "./connected";

interface RunControls {
  readonly launching: boolean;
  readonly button: ComponentProps<typeof RunButton>["button"];
  readonly toggle: () => void;
}

function summary(state: ReadyPlant, hall: string | null): string {
  const ports = state.plant.links.some((link) => link.fromPort === null) ? " · some links record no ports" : "";
  const figures =
    hall === null
      ? `${counted(state.plant.instances.length, "instance", "instances")} · ${counted(state.plant.links.length, "link", "links")}`
      : hallCount(plantView(state.plant, state.specs, { open: hall }), hall);
  return `${figures}${ports}`;
}

function Crumbs({
  state,
  hall,
  navigation,
}: {
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly navigation: NavigationStore;
}) {
  const title = plantTitle(state.plant);
  const heading = useScreen(hall === null ? `${title} · plant` : `${hall} · ${title}`);
  const trail: Crumb[] = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "workspace" });
      },
    },
  ];
  if (hall !== null) {
    trail.push({
      label: title,
      onGo: () => {
        navigation.go({ screen: "plant", domain: state.path, hall: null });
      },
    });
  }
  return <Breadcrumb trail={trail} here={hall ?? title} heading={heading} />;
}

interface PlantCentreProps {
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly run: RunControls;
  readonly connected?: Connected;
}

function LiveViewButton({
  connected,
  navigation,
}: {
  readonly connected: Connected;
  readonly navigation: NavigationStore;
}) {
  return (
    <button
      type="button"
      className={shared.primary}
      onClick={() => {
        navigation.go({ screen: "review", worldId: connected.run });
      }}
    >
      <Radio size={14} aria-hidden="true" />
      Open live view
    </button>
  );
}

function Bar(props: PlantCentreProps & { readonly running: string | null }) {
  const { state, hall, navigation, run, connected, running } = props;
  return (
    <div className={screen.bar}>
      <Network size={14} aria-hidden="true" />
      <Crumbs state={state} hall={hall} navigation={navigation} />
      <span className={styles.summary}>{summary(state, hall)}</span>
      <EditButton composition={state.composition} running={running} navigation={navigation} />
      {connected?.kind === "session" && <LiveViewButton connected={connected} navigation={navigation} />}
      <RunButton
        launching={run.launching}
        running={running}
        button={run.button}
        onToggle={run.toggle}
        secondary={connected?.kind === "session"}
      />
      {connected?.trendsToggle}
      <ZoomBar />
    </div>
  );
}

const noFaults: readonly PlannedFault[] = [];

function usePlannedFaults(launch: LaunchStore, plant: string): readonly PlannedFault[] {
  const run = useStore(launch);
  const plan = run.phase === "running" && run.origin.domain === plant ? run.plan : null;
  return useMemo(() => (plan === null ? noFaults : shownFaults(plan)), [plan]);
}

export function PlantCentre(props: PlantCentreProps) {
  const { state, hall, plants, navigation, launch, connected } = props;
  const running = useRunningHere(launch, state.plant.id);
  const planned = usePlannedFaults(launch, state.plant.id);
  const bar = <Bar {...props} running={running} />;
  return (
    <section
      className={cx(screen.center, connected?.kind === "session" && styles.withInterrupts)}
      aria-label={`Plant ${state.plant.id}`}
    >
      {connected?.kind !== "session" ? (
        bar
      ) : (
        <div className={styles.top}>
          {bar}
          {connected.interrupts}
        </div>
      )}
      <div className={styles.work}>
        <PlantCanvas
          state={state}
          hall={hall}
          plants={plants}
          navigation={navigation}
          launch={launch}
          {...(connected === undefined ? {} : { connected })}
        />
        {connected?.trends?.((id) => {
          plants.select(id);
        }, planned)}
      </div>
    </section>
  );
}
