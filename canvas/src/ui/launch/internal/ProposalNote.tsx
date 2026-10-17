import type { Proposal } from "../../../app/launch";
import { panel } from "../../primitives";

function proposalWords(proposed: Proposal): string {
  const source = proposed.from === "" ? "the earlier run" : proposed.from;
  const copied = `Prefilled from ${source}: seed, duration, mode and faults. Change them if needed.`;
  return proposed.left.length === 0 ? copied : `${copied} Not copied: ${proposed.left.join("; ")}.`;
}

export function ProposalNote({ proposed }: { readonly proposed: Proposal | null }) {
  if (proposed === null || (proposed.from === "" && proposed.left.length === 0)) return null;
  return (
    <p className={panel.hint} role="note">
      {proposalWords(proposed)}
    </p>
  );
}
