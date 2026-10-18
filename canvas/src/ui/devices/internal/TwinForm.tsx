import { pairsOf } from "../../../app/devices";
import { renameTwin, retitleTwin, twinIssuesOn, twinNameLooksRight, twinNameRule } from "../../../model/devices";
import { cx, panel } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, TextBox } from "./Field";
import { KindCard } from "./KindCard";
import { choose, commit } from "./twinProps";
import type { TwinProps } from "./twinProps";

const twinScope = { scope: "twin", kind: "", key: "" } as const;

function TitleField({ twin, title, issues }: Pick<TwinProps, "twin" | "issues"> & { readonly title: string }) {
  return (
    <Field label="Title" issues={twinIssuesOn(issues, { ...twinScope, field: "title" })}>
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

function Naming({ twin, draft, issues }: TwinProps) {
  const { id, title } = draft.doc;
  const likely = id === "" || twinNameLooksRight(id);
  return (
    <section className={panel.group} aria-labelledby="twin-title">
      <h2 id="twin-title" className={styles.formTitle}>
        Twin
      </h2>
      <Field
        label="Twin name"
        hint={twinNameRule}
        caution={likely ? "" : "The binary is likely to refuse this name."}
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
      <TitleField twin={twin} title={title} issues={issues} />
    </section>
  );
}

export function TwinForm(props: TwinProps) {
  const kinds = props.draft.members.map((member) => member.kind);
  return (
    <div className={cx(panel.body, styles.form)}>
      <p className={panel.hint}>
        The twin is a simulated version of your devices. Choose what measures each reading and how each command changes
        things. The choices are saved as you go.
      </p>
      <Naming {...props} />
      {pairsOf(props.draft).map(({ kind, choice }) => (
        <KindCard key={kind.kind} {...props} kind={kind} choice={choice} kinds={kinds} />
      ))}
    </div>
  );
}
