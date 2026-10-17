import { cx, panel, shared } from "../../primitives";
import type { NavigationStore } from "../../../app/shell";
import styles from "./MissingRun.module.css";

export function MissingRun({
  worldId,
  navigation,
}: {
  readonly worldId: string;
  readonly navigation: NavigationStore;
}) {
  return (
    <div className={styles.missing} role="alert">
      <p>{`Run ${worldId} is not in this workspace.`}</p>
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          navigation.go({ screen: "home" });
        }}
      >
        Back to the workspace
      </button>
    </div>
  );
}
