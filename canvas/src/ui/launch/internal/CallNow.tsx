import { useEffect, useRef, useState } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NowDraft } from "../../../model/launch";
import { emptyNowDraft, nowCallOf, nowCallProblem } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { cx, panel, shared } from "../../primitives";
import { CallArgs } from "./CallArgs";
import { EffectorPick } from "./CallRows";
import { InstancePick, ProblemNote } from "./RowParts";
import styles from "./RunControls.module.css";

const label = "Call now";
const id = "now-call";

function useFocusFirstProblem(tried: number) {
  const box = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (tried > 0) box.current?.querySelector<HTMLElement>("[aria-invalid='true']")?.focus();
  }, [tried]);
  return box;
}

interface CallNowProps {
  readonly plant: Plant;
  readonly launch: LaunchStore;
}

export function CallNow({ plant, launch }: CallNowProps) {
  const [draft, setDraft] = useState<NowDraft>(emptyNowDraft);
  const [tried, setTried] = useState(0);
  const box = useFocusFirstProblem(tried);
  const problem = tried > 0 ? nowCallProblem(draft, plant) : null;
  const part = (name: string) => (problem?.part === name ? problem.text : "");
  const update = (next: Partial<NowDraft>) => {
    setDraft((now) => ({ ...now, ...next }));
  };
  const call = () => {
    setTried((count) => count + 1);
    if (nowCallProblem(draft, plant) !== null) return;
    const { entity, effector, args } = nowCallOf(draft, plant);
    void launch.invokeNow(entity, effector, args);
  };
  return (
    <div ref={box} className={styles.actRow}>
      <InstancePick
        plant={plant}
        entity={draft.entity}
        label={label}
        id={id}
        problem={part("instance")}
        fresh={false}
        prompt="Choose an instance"
        onPick={(entity) => {
          update({ entity, effector: "", values: {} });
        }}
      />
      <EffectorPick plant={plant} call={draft} label={label} id={id} problem={part("effector")} update={update} />
      <CallArgs plant={plant} call={draft} label={label} id={id} problem={problem} update={update} />
      <ProblemNote id={id} text={problem?.text ?? ""} />
      <button type="button" className={cx(shared.secondary, panel.secondary)} onClick={call}>
        Call now
      </button>
    </div>
  );
}
