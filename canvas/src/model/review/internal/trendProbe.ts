import type { PlaybackIndex } from "./playback";
import type { Span } from "./trend";
import type { Watched } from "./watchList";

export type ProbeMove =
  { readonly kind: "move"; readonly atNs: number } | { readonly kind: "commit" } | { readonly kind: "clear" };

const stepFraction = 20;

function clamped(ns: number, span: Span): number {
  return Math.min(span.endNs, Math.max(span.startNs, ns));
}

function sampleStep(index: PlaybackIndex, watched: Watched, atNs: number, direction: 1 | -1): number | null {
  return index.neighbour(watched.entity, watched.channel, atNs, direction);
}

function moveBy(span: Span, atNs: number, direction: 1 | -1): ProbeMove {
  const step = (span.endNs - span.startNs) / stepFraction;
  return { kind: "move", atNs: clamped(atNs + direction * step, span) };
}

function moveToSample(next: number | null, span: Span): ProbeMove | null {
  return next === null ? null : { kind: "move", atNs: clamped(next, span) };
}

const commitKeys = new Set(["Enter", " "]);
const pagedKeys = new Set(["PageUp", "PageDown"]);
const forwardKeys = new Set(["ArrowRight", "PageDown"]);

interface Press {
  readonly key: string;
  readonly shift: boolean;
}

function fixedMove(span: Span, key: string): ProbeMove | null {
  if (key === "Home") return { kind: "move", atNs: span.startNs };
  if (key === "End") return { kind: "move", atNs: span.endNs };
  if (key === "Escape") return { kind: "clear" };
  return commitKeys.has(key) ? { kind: "commit" } : null;
}

export function probeKey(
  index: PlaybackIndex,
  watched: Watched,
  span: Span,
  atNs: number,
  press: Press,
): ProbeMove | null {
  const direction = forwardKeys.has(press.key) ? 1 : -1;
  const arrow = press.key === "ArrowLeft" || press.key === "ArrowRight";
  if (arrow && !press.shift) return moveToSample(sampleStep(index, watched, atNs, direction), span);
  if (arrow || pagedKeys.has(press.key)) return moveBy(span, atNs, direction);
  return fixedMove(span, press.key);
}
