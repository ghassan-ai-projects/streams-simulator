import { isMissing } from "../../model/bridge";
import type { BridgeFailure } from "../../model/bridge";
import type { NavigationStore } from "../../app/navigationStore";
import { Breadcrumb } from "../shell/Breadcrumb";
import panel from "../panel.module.css";
import { useScreen } from "../useScreen";
import styles from "./PlantScreen.module.css";

export function PlantFailure({
  failure,
  navigation,
}: {
  readonly failure: BridgeFailure;
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
      <div className={styles.failure} role="alert">
        <p>{isMissing(failure) ? "This plant is not in the workspace any more." : "This plant could not be opened."}</p>
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
          <p>{`${failure.message} (${failure.code})`}</p>
        </details>
      </div>
    </>
  );
}
