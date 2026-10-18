import { Info, TriangleAlert } from "lucide-react";
import type { ReviewState } from "../../../app/review";
import { readyReport } from "../../../app/review";
import type { BannerWarning } from "../../../model/review";
import { bannerWarnings } from "../../../model/review";
import styles from "./Notices.module.css";

function Warning({ warning }: { readonly warning: BannerWarning }) {
  const full = warning.details === "" ? warning.text : `${warning.text} ${warning.details}`;
  return (
    <span className={styles.warning} role="alert" title={full}>
      <TriangleAlert size={14} aria-hidden="true" />
      <span className={styles.line}>{warning.text}</span>
      {warning.details !== "" && <span className="sr-only">{warning.details}</span>}
    </span>
  );
}

const domainChange = "The plant file changed after this run. The review and the replay use the plant the run recorded.";

function DomainNote({ state }: { readonly state: ReviewState }) {
  if (state.verified.phase !== "ready" || !state.verified.value.domainChanged) return null;
  return (
    <span className={styles.info} role="status" title={domainChange}>
      <Info size={14} aria-hidden="true" />
      <span className={styles.line}>{domainChange}</span>
    </span>
  );
}

export function Notices({ state }: { readonly state: ReviewState }) {
  if (state.connection === "running") return null;
  const list = bannerWarnings(readyReport(state), state.verified, state.playback);
  return (
    <div className={styles.notices}>
      <DomainNote state={state} />
      {list.map((warning) => (
        <Warning key={warning.text} warning={warning} />
      ))}
    </div>
  );
}
