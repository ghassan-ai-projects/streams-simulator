import { heldIssues, heldWords, serialiseKind, stemOf } from "../../../model/devices";
import type { Member } from "../../../model/devices";
import type { DeviceFailure, DevicesGateway, Outcome } from "../../ports";
import type { Held } from "./twinDraft";

export interface Suggestion {
  readonly outcome: Outcome<Readonly<Record<string, unknown>>, DeviceFailure>;
  readonly held: readonly Held[];
}

function nameOf(member: Member): string {
  return member.kind.kind || stemOf(member.path);
}

async function heldOf(gateway: DevicesGateway, members: readonly Member[]): Promise<readonly Held[]> {
  const checks = await Promise.all(members.map((member) => gateway.checkContent(serialiseKind(member.kind), true)));
  return members.flatMap((member, index) => {
    const checked = checks[index];
    if (checked === undefined || checked.ok || checked.failure.kind !== "refused") return [];
    const words = heldWords(nameOf(member), checked.failure.issues);
    const issues = heldIssues(member.kind, checked.failure.issues);
    return [{ path: member.path, name: nameOf(member), words, issues, adding: false }];
  });
}

export async function suggestFor(gateway: DevicesGateway, members: readonly Member[]): Promise<Suggestion> {
  const outcome = await gateway.suggest(members.map((member) => stemOf(member.path)));
  const refused = !outcome.ok && outcome.failure.kind === "refused";
  return { outcome, held: refused ? await heldOf(gateway, members) : [] };
}

export function unreadableHeld(path: string): Held {
  const name = stemOf(path);
  return { path, name, words: `${name} could not be read`, issues: [], adding: true };
}
