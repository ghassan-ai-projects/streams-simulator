import { ChevronRight, RotateCcw } from "lucide-react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { ReviewState } from "../../app/reviewStore";
import { planOfRun } from "../../model/reportViews";
import shared from "../shared.module.css";
import { useScreen } from "../useScreen";
import { Status } from "./RunStatus";
import styles from "./RunScreen.module.css";

export interface HallProps {
  readonly hall: string | null;
  readonly onHall: (hall: string | null) => void;
}

function Crumb({ label, onClick }: { readonly label: string; readonly onClick: () => void }) {
  return (
    <>
      <button type="button" className={shared.crumb} onClick={onClick}>
        {label}
      </button>
      <ChevronRight size={14} aria-hidden="true" />
    </>
  );
}

interface HeaderProps extends HallProps {
  readonly state: ReviewState;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly worldId: string;
  readonly plant: { readonly title: string; readonly path: string | null };
}

function RunActions({ state, navigation, launch, plant }: Omit<HeaderProps, "worldId" | "hall" | "onHall">) {
  const report = state.report.phase === "ready" ? state.report.value : null;
  const path = plant.path;
  if (path === null) return null;
  return (
    <div className={styles.runActions}>
      <button
        type="button"
        className={styles.runAction}
        onClick={() => {
          navigation.go({ screen: "plant", domain: path, hall: null });
        }}
      >
        Open plant
      </button>
      {report !== null && (
        <button
          type="button"
          className={styles.runAction}
          onClick={() => {
            launch.propose(report.domainId, planOfRun(report));
            navigation.go({ screen: "plant", domain: path, hall: null, launch: true });
          }}
        >
          <RotateCcw size={14} aria-hidden="true" />
          Run again
        </button>
      )}
    </div>
  );
}

export function Header(props: HeaderProps) {
  const { state, navigation, worldId, hall, onHall, plant } = props;
  const run = `Run ${worldId}`;
  const heading = useScreen(hall === null ? `${run} · ${plant.title}` : `${hall} · ${run}`);
  const path = plant.path;
  return (
    <div className={styles.bar}>
      <nav className={styles.crumbs} aria-label="Breadcrumb">
        <Crumb
          label="Workspace"
          onClick={() => {
            navigation.go({ screen: "home" });
          }}
        />
        {path !== null && (
          <Crumb
            label={plant.title}
            onClick={() => {
              navigation.go({ screen: "plant", domain: path, hall: null });
            }}
          />
        )}
        {hall !== null && (
          <Crumb
            label={run}
            onClick={() => {
              onHall(null);
            }}
          />
        )}
        <h1 aria-current="location" className={styles.here} ref={heading} tabIndex={-1}>
          {hall ?? run}
        </h1>
      </nav>
      <Status state={state} />
      <RunActions {...props} />
    </div>
  );
}
