import type { PlantFailure as Failure } from "../../app/ports";
import type { NavigationStore } from "../../app/navigationStore";
import { Breadcrumb } from "../shell/Breadcrumb";
import panel from "../panel.module.css";
import { useScreen } from "../useScreen";
import styles from "./PlantScreen.module.css";

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
        navigation.go({ screen: "home" });
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

function FailureBody({
  failure,
  navigation,
}: {
  readonly failure: Failure;
  readonly navigation: NavigationStore;
}) {
  return (
    <div className={styles.failure} role="alert">
      <p>
        {failure.kind === "missing"
          ? "This plant is not in the workspace any more."
          : "This plant could not be opened."}
      </p>
      <button
        type="button"
        className={panel.secondary}
        onClick={() => {
          navigation.go({ screen: "home" });
        }}
      >
        Back to the workspace
      </button>
      <details className={styles.detail}>
        <summary>Details</summary>
        <p>{`${failure.diagnostic.message} (${failure.diagnostic.displayCode})`}</p>
      </details>
    </div>
  );
}
