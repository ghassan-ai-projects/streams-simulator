import { CircleHelp, CircuitBoard, SquareDashed } from "lucide-react";
import { cx } from "./cx";
import styles from "./OriginBadge.module.css";

const icons = { real: CircuitBoard, simulated: SquareDashed, unknown: CircleHelp } as const;

type OriginBadgeKind = keyof typeof icons;

interface OriginBadgeProps {
  readonly kind: OriginBadgeKind;
  readonly label: string;
  readonly title: string;
  readonly said: string;
  readonly variant: "tag" | "glyph";
}

export function OriginBadge({ kind, label, title, said, variant }: OriginBadgeProps) {
  const Icon = icons[kind];
  const tagged = variant === "tag";
  return (
    <span className={cx(styles.origin, styles[kind], tagged && styles.tag)} title={title} data-origin={kind}>
      <Icon size={12} aria-hidden="true" />
      {tagged && (
        <span aria-hidden="true" data-origin-word="true">
          {label}
        </span>
      )}
      {said !== "" && <span className="sr-only">{said}</span>}
    </span>
  );
}
