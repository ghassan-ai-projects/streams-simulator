import type { InstanceOrigin } from "../../../model/graph";
import { originLabels, originWords } from "../../../model/graph";
import { OriginBadge } from "../../primitives";
import marks from "./marks.module.css";

interface OriginMarkProps {
  readonly origin: InstanceOrigin | null;
  readonly variant?: "tag" | "glyph";
  readonly voice?: "spoken" | "silent";
}

export function OriginMark({ origin, variant = "tag", voice = "spoken" }: OriginMarkProps) {
  if (origin === null) return null;
  const words = originWords(origin);
  return (
    <span className={marks.origin}>
      <OriginBadge
        kind={origin.kind}
        label={originLabels[origin.kind]}
        title={words}
        said={voice === "spoken" ? words : ""}
        variant={variant}
      />
    </span>
  );
}
