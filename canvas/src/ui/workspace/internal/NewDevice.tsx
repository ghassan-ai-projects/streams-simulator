import type { SyntheticEvent } from "react";
import { useRef, useState } from "react";
import type { DeviceDesk } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { deviceNameProblem } from "../../../model/devices";
import { shared } from "../../primitives";
import { Field } from "./NewPlantField";
import styles from "./NewPlant.module.css";
import workspaceStyles from "./WorkspaceScreen.module.css";

interface NewDeviceProps {
  readonly taken: readonly string[];
  readonly desk: DeviceDesk;
  readonly navigation: NavigationStore;
  readonly onCancel: () => void;
}

function useCreateDevice({ desk, navigation }: Pick<NewDeviceProps, "desk" | "navigation">) {
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState("");
  const sending = useRef(false);
  const create = (name: string) => {
    if (sending.current) return;
    sending.current = true;
    setBusy(true);
    setFailure("");
    void desk.create(name).then((made) => {
      sending.current = false;
      setBusy(false);
      if (made.ok) navigation.go({ screen: "device", path: made.value });
      else setFailure(made.failure.diagnostic.message);
    });
  };
  return { busy, failure, create };
}

interface ActionsProps {
  readonly blocked: boolean;
  readonly reason: string;
  readonly onCancel: () => void;
}

function DeviceActions({ blocked, reason, onCancel }: ActionsProps) {
  return (
    <span className={styles.newActions}>
      <button type="submit" className={shared.primary} aria-disabled={blocked}>
        Create and open the device
      </button>
      <button type="button" className={shared.secondary} onClick={onCancel}>
        Cancel
      </button>
      <span className={workspaceStyles.meta}>{reason}</span>
    </span>
  );
}

function DeviceStatus({ busy, failure }: { readonly busy: boolean; readonly failure: string }) {
  return (
    <>
      {busy && (
        <p role="status" className={workspaceStyles.meta}>
          Creating the device…
        </p>
      )}
      {failure !== "" && (
        <p role="alert" className={styles.newFailure}>
          {failure}
        </p>
      )}
    </>
  );
}

export function NewDeviceForm(props: NewDeviceProps) {
  const [name, setName] = useState("");
  const { busy, failure, create } = useCreateDevice(props);
  const problem = deviceNameProblem(name, props.taken);
  const typed = name.trim() !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (problem === "" && !busy) create(name.trim());
  };
  return (
    <form
      className={styles.newPlant}
      aria-label="New device"
      noValidate
      onSubmit={submit}
      onKeyDown={(event) => {
        if (event.key === "Escape") props.onCancel();
      }}
    >
      <Field
        first
        label="Name"
        placeholder="e.g. Cooling fan"
        hint="Becomes the device's name and its file name."
        value={name}
        onChange={setName}
        problem={typed ? problem : ""}
      />
      <DeviceActions blocked={problem !== "" || busy} reason={typed ? "" : problem} onCancel={props.onCancel} />
      <DeviceStatus busy={busy} failure={failure} />
    </form>
  );
}
