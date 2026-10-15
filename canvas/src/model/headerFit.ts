import { textFloorPx } from "./viewport";

const glyph = 0.62;
const idScreenPx = 12.5;
const headerInner = 206;
const markGap = 8;
const markPad = 4;
const iconPx = 12;
const iconCap = 22;
const markCap = 0.55;

export interface FittedId {
  readonly text: string;
  readonly size: number;
}

export function middleCut(text: string, chars: number): string {
  if (text.length <= chars) return text;
  const keep = Math.max(2, chars - 1);
  const head = Math.ceil(keep * 0.4);
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

export function markRoom(zoom: number, wordsLength: number): number {
  const icon = Math.min(iconPx / zoom, iconCap);
  const chip = icon + (2 * markPad) / zoom;
  const font = Math.max(textFloorPx / zoom, Math.min(11 / zoom, iconCap));
  const words = wordsLength === 0 ? 0 : wordsLength * 0.55 * font + 3;
  return Math.min(chip + words, headerInner * markCap) + markGap;
}

export function idRoom(zoom: number, wordsLength: number | null): number {
  return headerInner - (wordsLength === null ? 0 : markRoom(zoom, wordsLength));
}
