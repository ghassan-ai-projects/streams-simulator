import { pairsOf } from "../../../app/devices";
import { renameTwin, retitleTwin, twinIssuesOn } from "../../../model/devices";
import { cx, panel } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, TextBox } from "./Field";
import { FormTitle } from "./FormTitle";
import { DeviceNeeds } from "./InterfaceNeeds";
import { KindCard } from "./KindCard";
import { choose, commit } from "./twinProps";
import type { TwinProps } from "./twinProps";

const twinScope = { scope: "twin", kind: "", key: "" } as const;

function Naming({ twin, draft, issues }: TwinProps) {
  const { id, title } = draft.doc;
  return (
    <section className={panel.group} aria-labelledby="twin-title">
      <FormTitle id="twin-title" words="Twin" />
      <Field
        label="Twin name"
        field="id"
        hint="The plant made from the twin is saved under this name."
        issues={twinIssuesOn(issues, { ...twinScope, field: "id" })}
      >
        {(aria) => (
          <TextBox
            aria={aria}
            value={id}
            onChange={(next) => {
              twin.edit((choices) => renameTwin(choices, next));
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
      <Field label="Title" field="title" issues={twinIssuesOn(issues, { ...twinScope, field: "title" })}>
        {(aria) => (
          <TextBox
            aria={aria}
            maxLength={120}
            value={title}
            onChange={(next) => {
              choose(twin, (choices) => retitleTwin(choices, next));
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
    </section>
  );
}

export function TwinForm(props: TwinProps) {
  return (
    <div className={cx(panel.body, styles.form)}>
      <p className={panel.hint}>
        The twin is a simulated version of your devices. Choose what measures each reading and how each command changes
        things. The choices are saved as you go.
      </p>
      <DeviceNeeds needs={props.needs} reveal={props.reveal} />
      <Naming {...props} />
      {pairsOf(props.draft).map(({ kind, choice }) => (
        <KindCard key={choice.kind} {...props} kind={kind} choice={choice} />
      ))}
    </div>
  );
}
