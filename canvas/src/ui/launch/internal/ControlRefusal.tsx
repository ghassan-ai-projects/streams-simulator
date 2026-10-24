import { TriangleAlert } from "lucide-react";
import styles from "./RunControls.module.css";

export function ControlRefusal({ words }: { readonly words: string }) {
  if (words === "") return null;
  return (
    <p className={styles.refusal} role="alert">
      <TriangleAlert size={14} aria-hidden="true" />
      {words}
    </p>
  );
}
