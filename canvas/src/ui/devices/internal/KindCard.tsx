import { useId } from "react";
import {
  askedCommands,
  catalogueKey,
  commandChoiceOf,
  metricChoiceOf,
  placeKey,
  renameInstance,
  twinIssuesOn,
} from "../../../model/devices";
import type { KindChoice, KindFile } from "../../../model/devices";
import { panel, shared } from "../../primitives";
import tw from "./Behaviour.module.css";
import { CommandRow } from "./CommandRow";
import styles from "./Devices.module.css";
import { Field, ItemIssues, TextBox } from "./Field";
import { ReadingRow } from "./ReadingRow";
import { commit } from "./twinProps";
import type { TwinProps } from "./twinProps";

interface CardProps extends TwinProps {
  readonly kind: KindFile;
  readonly choice: KindChoice;
}

function isOwn(draft: CardProps["draft"], choice: KindChoice): boolean {
  return draft.members[0]?.path === choice.kind;
}

function Readings({ twin, draft, issues, needs, kind, choice, reveal }: CardProps) {
  const named = kind.metrics.filter((metric) => metric.name !== "");
  if (named.length === 0) return <p className={panel.hint}>This device has no readings.</p>;
  return (
    <>
      {named.map((metric) => (
        <ReadingRow
          key={metric.key}
          twin={twin}
          reveal={isOwn(draft, choice) ? reveal : null}
          issues={issues}
          needs={isOwn(draft, choice) ? needs : []}
          member={choice.kind}
          metric={metric}
          choice={metricChoiceOf(choice, metric.key)}
          instruments={draft.instruments[catalogueKey(metric)] ?? null}
        />
      ))}
    </>
  );
}

function Commands({ twin, draft, issues, needs, kind, choice, reveal }: CardProps) {
  const asked = askedCommands(kind);
  const stops = kind.commands.filter((command) => command.safeStop);
  return (
    <>
      {asked.map((command) => (
        <CommandRow
          key={command.key}
          twin={twin}
          reveal={isOwn(draft, choice) ? reveal : null}
          issues={issues}
          needs={isOwn(draft, choice) ? needs : []}
          kind={kind}
          members={draft.members}
          member={choice.kind}
          command={command}
          choice={commandChoiceOf(choice, command.key)}
        />
      ))}
      {stops.length > 0 && (
        <p className={panel.hint}>
          {`${stops.map((command) => command.name).join(", ")}: a safe stop. The twin handles it itself.`}
        </p>
      )}
    </>
  );
}

export function KindCard(props: CardProps) {
  const { twin, issues, kind, choice } = props;
  const title = useId();
  const on = (field: string) => twinIssuesOn(issues, { scope: "kind", kind: choice.kind, key: "", field });
  return (
    <section className={tw.card} aria-labelledby={title}>
      <h2 id={title} className={styles.formTitle}>
        {kind.kind}
      </h2>
      <ItemIssues messages={[...on(""), ...on("kind"), ...on("metrics"), ...on("commands")]} />
      <Field
        label="Name in the twin"
        place={placeKey({ scope: "kind", kind: choice.kind, key: "", field: "instance" })}
        hint="Names this device inside the twin."
        issues={on("instance")}
      >
        {(aria) => (
          <TextBox
            aria={aria}
            value={choice.instance}
            onChange={(next) => {
              twin.edit((choices) => renameInstance(choices, choice.kind, next));
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
      <h3 className={shared.sectionTitle}>Readings</h3>
      <Readings {...props} />
      <h3 className={shared.sectionTitle}>Commands</h3>
      <Commands {...props} />
    </section>
  );
}
