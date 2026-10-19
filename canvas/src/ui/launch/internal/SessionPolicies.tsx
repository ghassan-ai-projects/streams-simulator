import type { ConnectSession, PolicyDefaultState } from "../../../app/launch";
import { policyListWords, saveDefaultHint, saveDefaultWords } from "../../../model/launch";
import { shortDigest } from "../../../model/shared/words";
import { cx, panel, shared, useStore } from "../../primitives";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

export function PolicyInEffect({ policies }: { readonly policies: readonly string[] }) {
  return (
    <section className={panel.group} aria-label="Allowed policies in effect">
      <p className={launch.status}>{policyListWords(policies)}</p>
      {policies.length > 0 && (
        <ul className={styles.digests} aria-label="Digests in effect">
          {policies.map((digest) => (
            <li key={digest}>
              <code aria-hidden="true" title={digest}>
                {shortDigest(digest)}
              </code>
              <span className="sr-only">{digest}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function SaveStatus({ defaults }: { readonly defaults: PolicyDefaultState }) {
  const { save, problem, policies } = defaults;
  if (save === "idle") return null;
  const refused = save === "conflict" || save === "failed";
  return (
    <p role={refused ? "alert" : "status"} className={refused ? launch.problem : panel.hint}>
      {saveDefaultWords(save, problem, policies)}
    </p>
  );
}

export function SaveDefault({ connect, domain }: { readonly connect: ConnectSession; readonly domain: string }) {
  const defaults = useStore(connect.defaults);
  const saving = defaults.save === "saving";
  return (
    <div className={styles.saveDefault}>
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        aria-disabled={saving}
        aria-describedby="session-default-hint"
        onClick={() => {
          if (!saving) void connect.saveDefault(domain);
        }}
      >
        Save as default
      </button>
      <small id="session-default-hint" className={panel.hint}>
        {saveDefaultHint}
      </small>
      <SaveStatus defaults={defaults} />
    </div>
  );
}
