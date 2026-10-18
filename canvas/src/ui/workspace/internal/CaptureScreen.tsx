import { useEffect, useState } from "react";
import type { CaptureSession, CaptureState, DeviceDesk } from "../../../app/devices";
import { decisionCounts, reviewedOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import type { Workspace } from "../../../model/workspace";
import { counted } from "../../../model/workspace";
import { Breadcrumb, useScreen, useStore } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureRefused } from "./CaptureRefused";
import { CaptureReview } from "./CaptureReview";
import { CaptureSource } from "./CaptureSource";

interface ScreenProps {
  readonly workspace: Workspace;
  readonly desk: DeviceDesk;
  readonly navigation: NavigationStore;
  readonly onClose: () => void;
}

function useCaptureSession(desk: DeviceDesk): CaptureSession {
  const [session] = useState(() => desk.capture());
  useEffect(
    () => () => {
      session.close();
    },
    [session],
  );
  return session;
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

interface ResultProps extends Pick<ScreenProps, "navigation"> {
  readonly state: CaptureState;
  readonly session: CaptureSession;
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

function useEscapeToClose(onClose: () => void) {
  useEffect(() => {
    const leave = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !event.defaultPrevented) onClose();
    };
    document.addEventListener("keydown", leave);
    return () => {
      document.removeEventListener("keydown", leave);
    };
  }, [onClose]);
}

export function CaptureScreen({ workspace, desk, navigation, onClose }: ScreenProps) {
  const session = useCaptureSession(desk);
  const state = useStore(session);
  const heading = useScreen("Capture from a board");
  useEscapeToClose(onClose);
  return (
    <>
      <Breadcrumb trail={[{ label: "Workspace", onGo: onClose }]} here="Capture from a board" heading={heading} />
      <div className={styles.page}>
        <p className={styles.lead}>
          Read what a connected board says it is, then decide what to keep. Nothing is written until you save a device.
        </p>
        <CaptureSource
          workspace={workspace}
          busy={state.phase === "reading" || state.phase === "saving"}
          onRead={(source) => {
            void session.read(source, workspace.kinds);
          }}
        />
        <p className={state.phase === "reading" ? styles.note : "sr-only"} role="status">
          {statusWords(state)}
        </p>
        <Result state={state} session={session} navigation={navigation} />
      </div>
    </>
  );
}
