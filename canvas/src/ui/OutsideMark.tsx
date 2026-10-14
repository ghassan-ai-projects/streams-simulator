import { CircleSlash } from "lucide-react";
import styles from "./OutsideMark.module.css";

export function OutsideMark({ text }: { readonly text: string }) {
  if (text === "") return null;
  return (
    <span className={styles.mark} role="img" aria-label={text} title={text}>
      <CircleSlash size={11} aria-hidden="true" />
    </span>
  );
}
