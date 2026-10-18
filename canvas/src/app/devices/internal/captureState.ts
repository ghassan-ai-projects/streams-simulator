import type { BoardCapture, CapturedKind, CaptureSource, DeviceFailure, FileContent, Outcome } from "../../ports";

export type Decision =
  | { readonly kind: "open" }
  | { readonly kind: "saving" }
  | { readonly kind: "saved"; readonly path: string }
  | { readonly kind: "skipped" }
  | { readonly kind: "failed"; readonly failure: DeviceFailure };

export type SavedFile = Outcome<FileContent, DeviceFailure> | null;

export interface Card {
  readonly kind: CapturedKind;
  readonly saved: SavedFile;
  readonly decision: Decision;
}

export interface Reviewed {
  readonly phase: "reviewing" | "saving" | "saved";
  readonly source: CaptureSource;
  readonly environment: BoardCapture["environment"];
  readonly bdSeq: number;
  readonly scan: Pick<BoardCapture, "birthsSeen" | "taken" | "laterBirths" | "issues">;
  readonly cards: readonly Card[];
}

export type CaptureState =
  | { readonly phase: "idle" }
  | { readonly phase: "reading"; readonly source: CaptureSource }
  | { readonly phase: "refused"; readonly source: CaptureSource; readonly failure: DeviceFailure }
  | Reviewed;

function settled(card: Card): boolean {
  const { kind } = card.decision;
  return kind === "saved" || kind === "skipped" || card.kind.comparison.status === "same";
}

function decided(card: Card): boolean {
  return card.decision.kind === "saved" || card.decision.kind === "skipped";
}

function phaseOf(cards: readonly Card[]): Reviewed["phase"] {
  if (cards.some((card) => card.decision.kind === "saving")) return "saving";
  return cards.every(settled) && cards.some(decided) ? "saved" : "reviewing";
}

export function reviewed(source: CaptureSource, capture: BoardCapture, saved: readonly SavedFile[]): Reviewed {
  const cards = capture.kinds.map((kind, index): Card => ({
    kind,
    saved: saved[index] ?? null,
    decision: { kind: "open" },
  }));
  const { environment, bdSeq, birthsSeen, taken, laterBirths, issues } = capture;
  const scan = { birthsSeen, taken, laterBirths, issues };
  return { phase: phaseOf(cards), source, environment, bdSeq, scan, cards };
}

export function withDecision(state: Reviewed, index: number, decision: Decision): Reviewed {
  const cards = state.cards.map((card, at) => (at === index ? { ...card, decision } : card));
  return { ...state, cards, phase: phaseOf(cards) };
}
