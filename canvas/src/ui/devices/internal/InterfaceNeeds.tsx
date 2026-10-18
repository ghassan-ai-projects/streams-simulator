import type { FieldIssue, Selection } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import { ItemNeeds } from "./Field";
import type { Reveal } from "./twinProps";

interface NeedsProps {
  readonly needs: readonly FieldIssue[];
  readonly part: "metric" | "command";
  readonly itemKey: string;
  readonly reveal: Reveal | null;
}

interface BlockProps {
  readonly own: readonly FieldIssue[];
  readonly selection: Selection;
  readonly reveal: Reveal | null;
}

function Block({ own, selection, reveal }: BlockProps) {
  const first = own[0];
  return (
    <>
      <ItemNeeds messages={own.map((issue) => issue.text)} />
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

export function InterfaceNeeds({ needs, part, itemKey, reveal }: NeedsProps) {
  const own = needs.filter((issue) => issue.part === part && issue.key === itemKey);
  return <Block own={own} selection={{ part, key: itemKey }} reveal={reveal} />;
}

export function DeviceNeeds({ needs, reveal }: Pick<NeedsProps, "needs" | "reveal">) {
  const own = needs.filter((issue) => issue.part === "kind" || issue.part === "file");
  return <Block own={own} selection={{ part: "kind" }} reveal={reveal} />;
}
