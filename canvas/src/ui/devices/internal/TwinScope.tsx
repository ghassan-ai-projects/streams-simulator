import { scopeOf } from "../../../app/devices";
import type { Exclusion, TwinDraft } from "../../../app/devices";
import prov from "./Provenance.module.css";

function Lines({ lines, off }: { readonly lines: readonly string[]; readonly off: boolean }) {
  if (lines.length === 0) return <dd>None</dd>;
  return (
    <dd className={off ? prov.scopeOff : undefined}>
      {lines.map((line) => (
        <span key={line} className={prov.scopeLine}>
          {line}
        </span>
      ))}
    </dd>
  );
}

function reasons(list: readonly Exclusion[]): readonly string[] {
  return list.map(({ name, reason }) => `${name}: ${reason}`);
}

export function TwinScope({ draft }: { readonly draft: TwinDraft }) {
  const { included, excluded, blocked } = scopeOf(draft);
  return (
    <section className={prov.scope} aria-label="What goes into this twin">
      <dl className={prov.scopeList}>
        <dt>Included</dt>
        <Lines lines={included} off={false} />
        <dt>Excluded</dt>
        <Lines lines={reasons(excluded)} off />
        <dt>Blocked</dt>
        <Lines lines={reasons(blocked)} off />
      </dl>
    </section>
  );
}
