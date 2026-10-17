import type { ComponentProps } from "react";
import { Network } from "lucide-react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantSession, PlantState } from "../../app/routeModules";
import { hallCount, plantView } from "../../model/graph";
import { counted } from "../../model/workspace";
import type { Crumb } from "../shell/Breadcrumb";
import { Breadcrumb } from "../shell/Breadcrumb";
import { useScreen } from "../useScreen";
import { ZoomBar } from "../graph";
import { EditButton, RunButton, useRunningHere } from "./PlantActions";
import { PlantCanvas } from "./PlantCanvas";
import styles from "./PlantScreen.module.css";

export type Ready = Extract<PlantState, { phase: "ready" }>;

export interface RunControls {
  readonly launching: boolean;
  readonly button: ComponentProps<typeof RunButton>["button"];
  readonly toggle: () => void;
}

function summary(state: Ready, hall: string | null): string {
  const ports = state.plant.links.some((link) => link.fromPort === null) ? " · some links record no ports" : "";
  const figures =
    hall === null
      ? `${counted(state.plant.instances.length, "instance", "instances")} · ${counted(state.plant.links.length, "link", "links")}`
      : hallCount(plantView(state.plant, state.specs, hall), hall);
  return `${figures}${ports}`;
}

function Crumbs({
  state,
  hall,
  navigation,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly navigation: NavigationStore;
}) {
  const title = state.plant.title || state.plant.id;
  const heading = useScreen(hall === null ? `${title} · plant` : `${hall} · ${title}`);
  const trail: Crumb[] = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "home" });
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

export function PlantCentre({
  state,
  hall,
  plants,
  navigation,
  launch,
  run,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly run: RunControls;
}) {
  const running = useRunningHere(launch, state.plant.id);
  return (
    <section className={styles.center} aria-label={`Plant ${state.plant.id}`}>
      <div className={styles.bar}>
        <Network size={14} aria-hidden="true" />
        <Crumbs state={state} hall={hall} navigation={navigation} />
        <span className={styles.summary}>{summary(state, hall)}</span>
        <EditButton composition={state.composition} running={running} navigation={navigation} />
        <RunButton launching={run.launching} running={running} button={run.button} onToggle={run.toggle} />
        <ZoomBar />
      </div>
      <PlantCanvas state={state} hall={hall} plants={plants} navigation={navigation} launch={launch} />
    </section>
  );
}
