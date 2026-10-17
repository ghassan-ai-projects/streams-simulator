import { shared, cx } from "../../primitives";
import { RotateCcw } from "lucide-react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { ReviewState } from "../../../app/review";
import { rerunOf } from "../../../model/review";
import type { Crumb } from "../../primitives";
import { Breadcrumb } from "../../primitives";
import { useScreen } from "../../primitives";
import { Status } from "./RunStatus";
import styles from "./RunScreen.module.css";

export interface HallProps {
  readonly hall: string | null;
  readonly onHall: (hall: string | null) => void;
}

interface HeaderProps extends HallProps {
  readonly state: ReviewState;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly worldId: string;
  readonly plant: { readonly title: string; readonly path: string | null };
}

function RunActions({ state, navigation, launch, plant, worldId }: Omit<HeaderProps, "hall" | "onHall">) {
  const report = state.report.phase === "ready" ? state.report.value : null;
  const path = plant.path;
  if (path === null) return null;
  return (
    <div className={styles.runActions}>
      <button
        type="button"
        className={cx(styles.runAction, shared.button)}
        onClick={() => {
          navigation.go({ screen: "plant", domain: path, hall: null });
        }}
      >
        Open plant
      </button>
      {report !== null && (
        <button
          type="button"
          className={cx(styles.runAction, shared.button)}
          onClick={() => {
            const rerun = rerunOf(report);
            launch.propose(report.domainId, rerun.plan, rerun.left, worldId);
            navigation.go({ screen: "plant", domain: path, hall: null, launch: true });
          }}
        >
          <RotateCcw size={14} aria-hidden="true" />
          Set up another run
        </button>
      )}
    </div>
  );
}

function trailOf({ navigation, worldId, hall, onHall, plant }: HeaderProps): Crumb[] {
  const trail: Crumb[] = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "home" });
      },
    },
  ];
  const path = plant.path;
  if (path !== null) {
    trail.push({
      label: plant.title,
      onGo: () => {
        navigation.go({ screen: "plant", domain: path, hall: null });
      },
    });
  }
  if (hall !== null)
    trail.push({
      label: `Run ${worldId}`,
      onGo: () => {
        onHall(null);
      },
    });
  return trail;
}

export function Header(props: HeaderProps) {
  const { state, worldId, hall, plant } = props;
  const run = `Run ${worldId}`;
  const heading = useScreen(hall === null ? `${run} · ${plant.title}` : `${hall} · ${run}`);
  return (
    <div className={styles.bar}>
      <Breadcrumb trail={trailOf(props)} here={hall ?? run} heading={heading} />
      <Status state={state} />
      <RunActions {...props} />
    </div>
  );
}
