import styles from "./Devices.module.css";
import { useStepHeading } from "./stepFocus";

interface TitleProps {
  readonly id: string;
  readonly words: string;
  readonly interim?: boolean;
}

export function FormTitle({ id, words, interim = false }: TitleProps) {
  const heading = useStepHeading(interim ? "interim" : "final");
  return (
    <h2 id={id} ref={heading} tabIndex={-1} className={styles.formTitle}>
      {words}
    </h2>
  );
}
