import { Fragment } from "react";
import styles from "./shared.module.css";

interface MetaPartsProps {
  readonly parts: readonly string[];
}

export function MetaParts({ parts }: MetaPartsProps) {
  const shown = parts.filter((part) => part !== "");
  return shown.map((part, index) => (
    <Fragment key={`${String(index)}:${part}`}>
      <span className={styles.metaPart}>
        {part}
        {index < shown.length - 1 ? " ·" : ""}
      </span>
      {index < shown.length - 1 ? " " : null}
    </Fragment>
  ));
}
