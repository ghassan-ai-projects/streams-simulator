import { pairsOf } from "../../../app/devices";
import { placeKey, renameTwin, retitleTwin, twinIssuesOn } from "../../../model/devices";
import { cx, panel } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, TextBox } from "./Field";
import { FormTitle } from "./FormTitle";
import { HeldNotice } from "./HeldNotice";
import { DeviceNeeds } from "./InterfaceNeeds";
import { KindCard } from "./KindCard";
import { SuggestionNote } from "./Suggestions";
import { TwinScope } from "./TwinScope";
import { choose, commit } from "./twinProps";
import type { TwinProps } from "./twinProps";

const twinScope = { scope: "twin", kind: "", key: "" } as const;

function placeOf(field: string): string {
  return placeKey({ ...twinScope, field });
}

function Naming(props: TwinProps) {
  return (
    <section className={panel.group} aria-labelledby="twin-title">
      <FormTitle id="twin-title" words="Twin" />
      <NameField {...props} />
      <TitleField {...props} />
    </section>
  );
}

function NameField({ twin, draft, issues }: TwinProps) {
  const { id } = draft.doc;
  return (
    <Field
      label="Twin name"
      field="id"
      place={placeOf("id")}
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
  );
}

function TitleField({ twin, draft, issues }: TwinProps) {
  const { title } = draft.doc;
  return (
    <Field
      label="Title"
      field="title"
      place={placeOf("title")}
      issues={twinIssuesOn(issues, { ...twinScope, field: "title" })}
    >
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
  );
}

export function TwinForm(props: TwinProps) {
  const lead = props.draft.members[0]?.path;
  const others = props.draft.held.filter((entry) => entry.path !== lead);
  return (
    <div className={cx(panel.body, styles.form)}>
      <p className={panel.hint}>
        The twin is a simulated version of your devices. Choose what measures each reading and how each command changes
        things. The choices are saved as you go.
      </p>
      <TwinScope draft={props.draft} />
      <HeldNotice
        held={others}
        twin={props.twin}
        navigation={props.navigation}
        lead={lead ?? ""}
        reveal={props.reveal}
      />
      <SuggestionNote draft={props.draft} twin={props.twin} />
      <DeviceNeeds needs={props.needs} reveal={props.reveal} />
      <Naming {...props} />
      {pairsOf(props.draft).map(({ kind, choice }) => (
        <KindCard key={choice.kind} {...props} kind={kind} choice={choice} />
      ))}
    </div>
  );
}
