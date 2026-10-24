import { useState } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { Plant } from "../../../model/plant";
import { faultsOf } from "../../../model/plant";
import { humanise } from "../../../model/shared/words";
import { cx, panel, shared } from "../../primitives";
import { InstancePick } from "./RowParts";
import { controllable } from "./RunControls";
import type { Running } from "./RunControls";
import launchStyles from "./LaunchPanel.module.css";
import styles from "./RunControls.module.css";

const label = "Fault now";

interface FaultNowProps {
  readonly plant: Plant;
  readonly launch: LaunchStore;
  readonly state: Running;
}

interface FaultPickProps {
  readonly plant: Plant;
  readonly entity: string;
  readonly fault: string;
  readonly onPick: (fault: string) => void;
}

function FaultPick({ plant, entity, fault, onPick }: FaultPickProps) {
  return (
    <span className={launchStyles.sentence}>
      has a
      <select
        aria-label={`${label} fault`}
        value={fault}
        disabled={entity === ""}
        onChange={(event) => {
          onPick(event.target.value);
        }}
      >
        <option value="">Choose a fault</option>
        {(faultsOf(plant, entity) ?? []).map((name) => (
          <option key={name} value={name}>
            {humanise(name)}
          </option>
        ))}
      </select>
    </span>
  );
}

export function FaultNow({ plant, launch, state }: FaultNowProps) {
  const [entity, setEntity] = useState("");
  const [fault, setFault] = useState("");
  const ready = controllable(state) && entity !== "" && fault !== "";
  return (
    <div className={styles.actRow}>
      <InstancePick
        plant={plant}
        entity={entity}
        label={label}
        id="now-fault"
        problem=""
        fresh={false}
        prompt="Choose an instance"
        onPick={(picked) => {
          setEntity(picked);
          setFault("");
        }}
      />
      <FaultPick plant={plant} entity={entity} fault={fault} onPick={setFault} />
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        aria-disabled={!ready}
        onClick={() => {
          if (ready) void launch.injectNow(entity, fault);
        }}
      >
        Inject now
      </button>
    </div>
  );
}
