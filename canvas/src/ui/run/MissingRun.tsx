import type { NavigationStore } from "../../app/navigationStore";
import panel from "../panel.module.css";
import styles from "./RunScreen.module.css";

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
        className={panel.secondary}
        onClick={() => {
          navigation.go({ screen: "home" });
        }}
      >
        Back to the workspace
      </button>
    </div>
  );
}
