import { capturePath } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { BoardCapture, CapturedKind, CaptureSource, DeviceFailure } from "../../ports";

export type Decision =
  | { readonly kind: "open" }
  | { readonly kind: "saving" }
  | { readonly kind: "saved"; readonly path: string }
  | { readonly kind: "skipped" }
  | { readonly kind: "failed"; readonly failure: DeviceFailure };

export interface Card {
  readonly kind: CapturedKind;
  readonly path: string;
  readonly decision: Decision;
}

export interface Reviewed {
  readonly phase: "reviewing" | "saving" | "saved";
  readonly source: CaptureSource;
  readonly environment: BoardCapture["environment"];
  readonly bdSeq: number;
  readonly cards: readonly Card[];
}

export type CaptureState =
  | { readonly phase: "idle" }
  | { readonly phase: "reading"; readonly source: CaptureSource }
  | { readonly phase: "refused"; readonly source: CaptureSource; readonly failure: DeviceFailure }
  | Reviewed;

function phaseOf(cards: readonly Card[]): Reviewed["phase"] {
  if (cards.some((card) => card.decision.kind === "saving")) return "saving";
  const decided = cards.every((card) => card.decision.kind === "saved" || card.decision.kind === "skipped");
  return cards.length > 0 && decided ? "saved" : "reviewing";
}

export function reviewed(source: CaptureSource, capture: BoardCapture, kinds: readonly KindEntry[]): Reviewed {
  const cards = capture.kinds.map((kind): Card => ({
    kind,
    path: capturePath(kinds, kind.kind),
    decision: { kind: "open" },
  }));
  return { phase: phaseOf(cards), source, environment: capture.environment, bdSeq: capture.bdSeq, cards };
}

export function withDecision(state: Reviewed, name: string, decision: Decision): Reviewed {
  const cards = state.cards.map((card) => (card.kind.kind === name ? { ...card, decision } : card));
  return { ...state, cards, phase: phaseOf(cards) };
}
