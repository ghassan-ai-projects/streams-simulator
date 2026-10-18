import type { OriginKind } from "../../shared/origin";
import { originLabels } from "../../shared/origin";

export { originLabels };

export interface InstanceOrigin {
  readonly instance: string;
  readonly kind: OriginKind;
  readonly board: string;
  readonly device: string;
  readonly bootId: string;
}

export function unknownOrigin(instance: string): InstanceOrigin {
  return { instance, kind: "unknown", board: "", device: "", bootId: "" };
}

export function originWords(origin: InstanceOrigin): string {
  if (origin.kind === "unknown") return `${originLabels.unknown}: the run recorded no origin for it`;
  const board = `${originLabels[origin.kind]} · board ${origin.board}`;
  return origin.kind === "real" && origin.bootId !== "" ? `${board} · boot ${origin.bootId}` : board;
}
