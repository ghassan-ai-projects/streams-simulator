import { useEffect, useState } from "react";
import type { CaptureSession, CaptureState, DeviceDesk } from "../../../app/devices";
import { decisionCounts, reviewedOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import type { WorkspaceState, WorkspaceStore } from "../../../app/workspace";
import type { Workspace } from "../../../model/workspace";
import { counted } from "../../../model/workspace";
import { Breadcrumb, screen, useScreen, useStore } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureRefused } from "./CaptureRefused";
import { CaptureReview } from "./CaptureReview";
import { CaptureSource } from "./CaptureSource";

interface ScreenProps {
  readonly workspace: WorkspaceStore;
  readonly desk: DeviceDesk;
  readonly navigation: NavigationStore;
}

interface ResultProps extends Pick<ScreenProps, "navigation"> {
  readonly state: CaptureState;
  readonly session: CaptureSession;
}

function useShownWorkspace(state: WorkspaceState): Workspace | null {
  const [last, setLast] = useState<Workspace | null>(null);
  if (state.phase === "ready" && state.workspace !== last) setLast(state.workspace);
  return state.phase === "ready" ? state.workspace : last;
}

function statusWords(state: CaptureState): string {
  if (state.phase === "reading") return "Reading the announcement…";
  const board = reviewedOf(state);
  if (board === null) return "";
  const { saved, skipped } = decisionCounts(board.cards);
  const decided = `${String(saved)} saved, ${String(skipped)} skipped.`;
  if (board.phase === "saved") return `All decided: ${decided}`;
  if (board.phase === "saving") return "Saving the device…";
  const announced = `The board announced ${counted(board.cards.length, "device")}.`;
  return saved + skipped === 0 ? announced : `${announced} ${decided}`;
}

function Result({ state, session, navigation }: ResultProps) {
  const board = reviewedOf(state);
  if (state.phase === "refused") {
    return (
      <CaptureRefused
        state={state}
        onRetry={() => {
          void session.retry();
        }}
      />
    );
  }
  return board === null ? null : <CaptureReview board={board} session={session} navigation={navigation} />;
}

function useEscapeToClose(onClose: () => void, active: boolean) {
  useEffect(() => {
    if (!active) return;
    const leave = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !event.defaultPrevented) onClose();
    };
    document.addEventListener("keydown", leave);
    return () => {
      document.removeEventListener("keydown", leave);
    };
  }, [onClose, active]);
}

export function CaptureScreen({ workspace, desk, navigation }: ScreenProps) {
  const session = desk.capture();
  const state = useStore(session);
  const shown = useShownWorkspace(useStore(workspace));
  const heading = useScreen("Capture from a board");
  const busy = state.phase === "reading" || state.phase === "saving";
  const leave = () => {
    session.leave();
    navigation.go({ screen: "workspace" });
  };
  useEscapeToClose(leave, !busy);
  return (
    <>
      <Breadcrumb trail={[{ label: "Workspace", onGo: leave }]} here="Capture from a board" heading={heading} />
      <div className={styles.page}>
        {shown === null ? (
          <p className={screen.loading} role="status">
            Loading the workspace…
          </p>
        ) : (
          <CaptureSource
            workspace={shown}
            busy={busy}
            read={reviewedOf(state)?.source ?? null}
            onRead={(source) => {
              void session.read(source);
            }}
            onClear={() => {
              session.clear();
            }}
          />
        )}
        <p className={state.phase === "reading" ? styles.note : "sr-only"} role="status">
          {statusWords(state)}
        </p>
        <Result state={state} session={session} navigation={navigation} />
      </div>
    </>
  );
}
