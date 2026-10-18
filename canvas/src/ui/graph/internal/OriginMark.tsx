import { CircuitBoard, SquareDashed } from "lucide-react";
import type { InstanceOrigin } from "../../../model/graph";
import { originLabels, originWords } from "../../../model/graph";
import { cx } from "../../primitives";
import styles from "./origin.module.css";

const icons = { real: CircuitBoard, simulated: SquareDashed } as const;

type OriginVariant = "tag" | "glyph";

export function OriginMark({
  origin,
  variant = "tag",
}: {
  readonly origin: InstanceOrigin | null;
  readonly variant?: OriginVariant;
}) {
  if (origin === null) return null;
  const Icon = icons[origin.kind];
  const tagged = variant === "tag";
  return (
    <span
      className={cx(styles.origin, styles[`origin-${origin.kind}`], tagged && styles.tag)}
      title={originWords(origin)}
      data-origin={origin.kind}
    >
      <Icon size={12} aria-hidden="true" />
      {tagged && <span aria-hidden="true">{originLabels[origin.kind]}</span>}
      <span className="sr-only">{originWords(origin)}</span>
    </span>
  );
}
