import styles from "./Devices.module.css";
import { useStepHeading } from "./stepFocus";

export function FormTitle({ id, words }: { readonly id: string; readonly words: string }) {
  const heading = useStepHeading();
  return (
    <h2 id={id} ref={heading} tabIndex={-1} className={styles.formTitle}>
      {words}
    </h2>
  );
}
