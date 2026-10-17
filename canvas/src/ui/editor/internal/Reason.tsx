import styles from "./Editor.module.css";

export function Reason({ id, text }: { readonly id: string; readonly text: string | null }) {
  if (text === null) return null;
  return (
    <p className={styles.reason} id={id}>
      {text}
    </p>
  );
}
