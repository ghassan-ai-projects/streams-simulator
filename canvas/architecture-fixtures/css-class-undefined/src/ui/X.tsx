import styles from "./x.module.css";

export function X() {
  return <p className={`${styles.known} ${styles.missing}`}>x</p>;
}
