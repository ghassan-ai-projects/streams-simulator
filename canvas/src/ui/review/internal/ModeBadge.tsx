import type { ReactNode } from "react";
import { ClipboardCheck, Hand, OctagonX } from "lucide-react";
import type { BadgeKind, InterruptMode } from "../../../model/review";
import { badgeKind, modeLine } from "../../../model/review";

interface ModeBadgeProps {
  readonly mode: InterruptMode;
  readonly latched: boolean;
  readonly size: number;
  readonly prefix?: string;
}

function iconOf(kind: BadgeKind, size: number): ReactNode {
  if (kind === "latched") return <OctagonX size={size} aria-hidden="true" />;
  return kind === "hold" ? <Hand size={size} aria-hidden="true" /> : <ClipboardCheck size={size} aria-hidden="true" />;
}

export function ModeBadge({ mode, latched, size, prefix = "" }: ModeBadgeProps) {
  return (
    <>
      {iconOf(badgeKind({ mode, latched }), size)}
      {`${prefix}${modeLine({ mode, latched })}`}
    </>
  );
}
