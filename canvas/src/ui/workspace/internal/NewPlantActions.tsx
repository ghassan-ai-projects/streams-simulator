import { shared } from "../../primitives";
import workspaceStyles from "./WorkspaceScreen.module.css";
import styles from "./NewPlant.module.css";

interface ActionsProps {
  readonly failure: string;
  readonly reason: string;
  readonly onCancel: () => void;
}

export function Actions({ failure, reason, onCancel }: ActionsProps) {
  const ready = reason === "";
  return (
    <>
      <span className={styles.newActions}>
        <button
          type="submit"
          className={shared.primary}
          aria-disabled={!ready}
          aria-describedby={ready ? undefined : "new-plant-reason"}
        >
          Create and open the editor
        </button>
        <button type="button" className={shared.secondary} onClick={onCancel}>
          Cancel
        </button>
        {!ready && (
          <span id="new-plant-reason" className={workspaceStyles.meta}>
            {reason}
          </span>
        )}
      </span>
      {failure !== "" && (
        <p role="alert" className={styles.newFailure}>
          {failure}
        </p>
      )}
    </>
  );
}
