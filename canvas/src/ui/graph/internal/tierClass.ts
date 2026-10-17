import type { tierOf } from "../../../model/graph";
import { cx } from "../../primitives";
import frames from "./frames.module.css";
import marks from "./marks.module.css";
import styles from "./nodes.module.css";
import pins from "./pins.module.css";
import readout from "./readout.module.css";

export const tierClass: Record<ReturnType<typeof tierOf>, string> = {
  map: cx(styles.tierMap, pins.tierMap, frames.tierMap, readout.tierMap, marks.tierMap),
  overview: cx(styles.tierOverview, pins.tierOverview, frames.tierOverview, readout.tierOverview, marks.tierOverview),
  compact: cx(styles.tierCompact, pins.tierCompact, frames.tierCompact, readout.tierCompact, marks.tierCompact),
  full: cx(styles.tierFull),
};
