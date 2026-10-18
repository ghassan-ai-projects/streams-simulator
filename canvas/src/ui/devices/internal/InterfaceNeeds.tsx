import { interfaceNeeds } from "../../../model/devices";
import type { Selection, TwinIssue } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import { ItemIssues } from "./Field";
import type { Reveal } from "./twinProps";

interface NeedsProps {
  readonly issues: readonly TwinIssue[];
  readonly scope: "metric" | "command";
  readonly kind: string;
  readonly name: string;
  readonly selection: Selection;
  readonly reveal: Reveal | null;
}

export function InterfaceNeeds({ issues, scope, kind, name, selection, reveal }: NeedsProps) {
  const needs = interfaceNeeds(issues, scope, kind, name);
  const first = needs[0];
  return (
    <>
      <ItemIssues messages={needs.map((issue) => issue.message)} />
      {reveal !== null && first !== undefined && (
        <button
          type="button"
          className={cx(shared.secondary, panel.secondary)}
          onClick={() => {
            reveal(selection, first.field);
          }}
        >
          Set it in the Interface step
        </button>
      )}
    </>
  );
}
