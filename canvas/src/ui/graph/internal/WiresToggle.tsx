import { Cable } from "lucide-react";
import { cx } from "../../primitives";
import type { WiresControl } from "./useWires";
import styles from "./GraphControls.module.css";

const hint =
  "Draw the links between two instances as one wire with their count, and list each instance's measured channels";

const linkingHint = " Switch it off to see the pins and draw a link between two of them.";

export function WiresToggle({ control, linking }: { readonly control: WiresControl; readonly linking: boolean }) {
  const grouped = control.value === "grouped";
  return (
    <button
      type="button"
      className={cx(styles.legendButton, grouped && styles.pressed)}
      aria-pressed={grouped}
      title={linking ? hint + linkingHint : hint}
      onClick={() => {
        control.onChange(grouped ? "each" : "grouped");
      }}
    >
      <Cable size={14} aria-hidden="true" />
      Group links
    </button>
  );
}
