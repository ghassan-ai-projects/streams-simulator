import type { FittedId } from "./types";
import type { IdTier } from "./viewport";
import { textFloorPx } from "./viewport";

const glyph = 0.605;
const idScreenPx = 12.5;
const headerInner = 206;
const nodeInner = 230;
const markGap = 8;
const markPad = 4;
const iconPx = 12;
const iconCap = 22;
const markCap = 0.55;
const wordsFontPx = 11;
const wordsGlyph = 0.55;
const wordsPad = 3;
const overviewIconCap = 40;
const overviewMarkPad = 8;
const overviewInset = 16;
const headShare = 0.4;

function middleCut(text: string, chars: number): string {
  if (text.length <= chars) return text;
  const keep = Math.max(2, chars - 1);
  const head = Math.ceil(keep * headShare);
  return `${text.slice(0, head)}…${text.slice(text.length - (keep - head))}`;
}

export function fitId(label: string, room: number, zoom: number, cap: number): FittedId {
  const wanted = Math.min(idScreenPx / zoom, cap);
  const floor = textFloorPx / zoom;
  const fits = room / (Math.max(1, label.length) * glyph);
  if (fits >= wanted) return { text: label, size: wanted };
  if (fits >= floor) return { text: label, size: fits };
  return { text: middleCut(label, Math.floor(room / (floor * glyph))), size: floor };
}

function markRoom(zoom: number, wordsLength: number): number {
  const icon = Math.min(iconPx / zoom, iconCap);
  const chip = icon + (2 * markPad) / zoom;
  const font = Math.max(textFloorPx / zoom, Math.min(wordsFontPx / zoom, iconCap));
  const words = wordsLength === 0 ? 0 : wordsLength * wordsGlyph * font + wordsPad;
  return Math.min(chip + words, headerInner * markCap) + markGap;
}

function overviewMark(zoom: number): number {
  return Math.min(iconPx / zoom, overviewIconCap) + overviewMarkPad / zoom;
}

interface IdTierFit {
  readonly cap: number;
  readonly inner: (zoom: number) => number;
  readonly marks: (zoom: number, wordsLength: number) => number;
}

const tierFits: Readonly<Record<IdTier, IdTierFit>> = {
  compact: { cap: 25, inner: () => headerInner, marks: markRoom },
  overview: { cap: 46, inner: (zoom) => nodeInner - overviewInset / zoom, marks: overviewMark },
};

export const idCap: Readonly<Record<IdTier, number>> = {
  compact: tierFits.compact.cap,
  overview: tierFits.overview.cap,
};

export function idRoom(zoom: number, wordsLength: number | null, tier: IdTier): number {
  const fit = tierFits[tier];
  return fit.inner(zoom) - (wordsLength === null ? 0 : fit.marks(zoom, wordsLength));
}
