import { Info } from "lucide-react";
import type { Reviewed } from "../../../app/devices";
import { scanNotices } from "../../../model/devices";
import styles from "./Capture.module.css";

export function CaptureNotices({ scan }: { readonly scan: Reviewed["scan"] }) {
  const notices = scanNotices(scan);
  if (notices.length === 0) return null;
  return (
    <section className={styles.notice} aria-label="About this announcement">
      {notices.map((words) => (
        <p key={words}>
          <Info size={14} aria-hidden="true" />
          {words}
        </p>
      ))}
      {scan.issues.length > 0 && (
        <details>
          <summary>Details</summary>
          {scan.issues.map((issue, index) => (
            <p key={`${String(index)}:${issue.message}`}>{issue.message}</p>
          ))}
        </details>
      )}
    </section>
  );
}
