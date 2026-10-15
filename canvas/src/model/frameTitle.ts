const type = {
  name: 13,
  line: 17,
  above: 4,
  cap: 30,
} as const;

export const capScale = type.cap / type.name;

export function countWords(count: number): string {
  return `${String(count)} ${count === 1 ? "node" : "nodes"}`;
}

export function titleScale(pixel: number): number {
  return Math.min(pixel, capScale);
}

export function titleBandHeight(scale: number): number {
  return type.line * scale;
}

export function titleAbove(scale: number): number {
  return type.above * scale;
}

export function headerReserve(): number {
  return titleBandHeight(capScale) + titleAbove(capScale);
}

export function boundsWithTitles(
  bounds: { readonly x: number; readonly y: number; readonly width: number; readonly height: number },
  titleTops: readonly number[],
  pixel: number,
) {
  const scale = titleScale(pixel);
  const tops = titleTops.map((y) => y - titleBandHeight(scale) - titleAbove(scale));
  const top = Math.min(bounds.y, ...tops);
  return { ...bounds, y: top, height: bounds.height + (bounds.y - top) };
}

const breaks = /[\s\-_/]/;
const tailChars = 6;

function tailStart(name: string): number {
  if (name.length <= tailChars + 2) return name.length;
  const lastBreak = name
    .split("")
    .map((char) => breaks.test(char))
    .lastIndexOf(true);
  const word = name.length - lastBreak - 1;
  return lastBreak > 0 && word <= tailChars ? lastBreak : name.length - tailChars;
}

export function splitName(name: string): { readonly head: string; readonly tail: string } {
  const at = tailStart(name);
  return { head: name.slice(0, at), tail: name.slice(at) };
}
