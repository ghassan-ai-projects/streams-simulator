import { cx, panel, screen, shared } from "../../primitives";
import type { PlantFailure as Failure } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import { failureLine } from "../../../model/shared/words";
import { Breadcrumb } from "../../primitives";
import { useScreen } from "../../primitives";
import styles from "./PlantFailure.module.css";

export function PlantFailure({
  failure,
  navigation,
}: {
  readonly failure: Failure;
  readonly navigation: NavigationStore;
}) {
  const heading = useScreen("Plant");
  const trail = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "workspace" });
      },
    },
  ];
  return (
    <>
      <Breadcrumb trail={trail} here="Plant" heading={heading} />
      <FailureBody failure={failure} navigation={navigation} />
    </>
  );
}

function FailureBody({ failure, navigation }: { readonly failure: Failure; readonly navigation: NavigationStore }) {
  return (
    <div className={screen.failure} role="alert">
      <p>
        {failure.kind === "missing"
          ? "This plant is not in the workspace any more."
          : "This plant could not be opened."}
      </p>
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          navigation.go({ screen: "workspace" });
        }}
      >
        Back to the workspace
      </button>
      <details className={styles.detail}>
        <summary>Details</summary>
        <p>{failureLine(failure.diagnostic)}</p>
      </details>
    </div>
  );
}
