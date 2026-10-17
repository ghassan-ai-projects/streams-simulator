import { counted } from "../../shared/words";
import type { LinkEnd } from "./composition";
import { instanceNameFits, instanceNameRule } from "./instanceNameRule";

export type LinkPart =
  | { readonly kind: "port"; readonly text: string }
  | { readonly kind: "arrow" }
  | { readonly kind: "other"; readonly id: string };

export function linkParts(end: LinkEnd): readonly LinkPart[] {
  const own: LinkPart = { kind: "port", text: end.port };
  const other: readonly LinkPart[] = [
    { kind: "other", id: end.other },
    { kind: "port", text: end.otherPort },
  ];
  const arrow: LinkPart = { kind: "arrow" };
  return end.outgoing ? [own, arrow, ...other] : [...other, arrow, own];
}

export function linkWords(end: LinkEnd): string {
  return end.outgoing ? `${end.port} to ${end.other} ${end.otherPort}` : `${end.other} ${end.otherPort} to ${end.port}`;
}

export function deleteQuestion(id: string, links: number): string {
  return `Delete ${id} and its ${counted(links, "link")}?`;
}

export function nameCheck(name: string): { readonly fits: boolean; readonly hint: string } {
  const fits = instanceNameFits(name);
  return { fits, hint: fits ? instanceNameRule() : `Not a valid name. ${instanceNameRule()}` };
}
