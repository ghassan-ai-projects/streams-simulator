import type { Proposal } from "../../../app/launch";
import { freshSettingsWords } from "../../../model/launch";
import { panel } from "../../primitives";

function proposalWords(proposed: Proposal): string {
  if (proposed.fresh) return freshSettingsWords(proposed.plan.explore);
  const source = proposed.from === "" ? "the earlier run" : proposed.from;
  const copied = `Prefilled from ${source}: name, purpose, seed, duration, mode, pace, faults with their ends, delivery conditions and effector calls. Change them if needed.`;
  return proposed.left.length === 0 ? copied : `${copied} Not copied: ${proposed.left.join("; ")}.`;
}

export function ProposalNote({ proposed }: { readonly proposed: Proposal | null }) {
  if (proposed === null || (proposed.from === "" && proposed.left.length === 0 && !proposed.fresh)) return null;
  return (
    <p className={panel.hint} role="note">
      {proposalWords(proposed)}
    </p>
  );
}
