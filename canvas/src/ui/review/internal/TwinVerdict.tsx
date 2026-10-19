import { CircleCheck, CircleQuestionMark, TriangleAlert } from "lucide-react";
import type { TwinVerdict } from "../../../model/review";
import { verdictWords } from "../../../model/review";
import { cx } from "../../primitives";
import chips from "./SessionEvidence.module.css";

const icons = { agrees: CircleCheck, diverges: TriangleAlert, cannot_judge: CircleQuestionMark } as const;

const tones: Readonly<Record<TwinVerdict, string | undefined>> = {
  agrees: chips.ok,
  diverges: chips.alarm,
  cannot_judge: undefined,
};

export function VerdictChip({ verdict }: { readonly verdict: TwinVerdict }) {
  const Icon = icons[verdict];
  return (
    <span className={cx(chips.chip, tones[verdict])}>
      <Icon size={13} aria-hidden="true" />
      {verdictWords[verdict].label}
    </span>
  );
}
