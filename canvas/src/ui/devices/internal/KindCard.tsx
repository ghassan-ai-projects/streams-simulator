import { useId } from "react";
import { askedCommands, commandChoiceOf, metricChoiceOf, renameInstance, twinIssuesOn } from "../../../model/devices";
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
  readonly kinds: readonly KindFile[];
}

function isOwn(draft: CardProps["draft"], kind: KindFile): boolean {
  return draft.members[0]?.kind.kind === kind.kind;
}

function Readings({ twin, draft, issues, kind, choice, reveal }: CardProps) {
  const named = kind.metrics.filter((metric) => metric.name !== "");
  if (named.length === 0) return <p className={panel.hint}>This device has no readings.</p>;
  return (
    <>
      {named.map((metric) => (
        <ReadingRow
          key={metric.key}
          twin={twin}
          reveal={isOwn(draft, kind) ? reveal : null}
          issues={issues}
          kind={kind.kind}
          metric={metric}
          choice={metricChoiceOf(choice, metric.name)}
          instruments={draft.instruments}
        />
      ))}
    </>
  );
}

function Commands({ twin, draft, issues, kind, kinds, choice, reveal }: CardProps) {
  const asked = askedCommands(kind);
  const stops = kind.commands.filter((command) => command.safeStop);
  return (
    <>
      {asked.map((command) => (
        <CommandRow
          key={command.key}
          twin={twin}
          reveal={isOwn(draft, kind) ? reveal : null}
          issues={issues}
          kind={kind}
          kinds={kinds}
          command={command}
          choice={commandChoiceOf(choice, command.name)}
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
  const on = (field: string) => twinIssuesOn(issues, { scope: "kind", kind: kind.kind, key: "", field });
  return (
    <section className={tw.card} aria-labelledby={title}>
      <h2 id={title} className={styles.formTitle}>
        {kind.kind}
      </h2>
      <ItemIssues messages={[...on(""), ...on("kind"), ...on("metrics"), ...on("commands")]} />
      <Field label="Name in the twin" hint="Names this device inside the twin." issues={on("instance")}>
        {(aria) => (
          <TextBox
            aria={aria}
            value={choice.instance}
            onChange={(next) => {
              twin.edit((choices) => renameInstance(choices, kind.kind, next));
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
