const type = {
  name: 13,
  count: 12,
  line: 17,
  above: 4,
  nameChar: 6.6,
  countChar: 5.9,
  gap: 8,
  pad: 4,
  cap: 30,
} as const;

export const capScale = type.cap / type.name;

export function countWords(count: number): string {
  return `${String(count)} ${count === 1 ? "node" : "nodes"}`;
}

export function titleScale(pixel: number): number {
  return Math.min(pixel, capScale);
}

export function titleLines(title: string, count: number, width: number, scale: number): number {
  const countWidth = (countWords(count).length * type.countChar + type.gap) * scale;
  const room = Math.max(1, width - countWidth);
  const need = (title.length * type.nameChar + type.pad) * scale;
  return Math.max(1, Math.ceil(need / room));
}

export function titleBandHeight(title: string, count: number, width: number, scale: number): number {
  return titleLines(title, count, width, scale) * type.line * scale;
}

export function titleAbove(scale: number): number {
  return type.above * scale;
}

export function headerReserve(title: string, count: number, width: number): number {
  return titleBandHeight(title, count, width, capScale) + titleAbove(capScale);
}

export interface PlacedTitle {
  readonly y: number;
  readonly width: number;
  readonly title: string;
  readonly count: number;
}

export function boundsWithTitles(
  bounds: { readonly x: number; readonly y: number; readonly width: number; readonly height: number },
  titles: readonly PlacedTitle[],
  pixel: number,
) {
  const scale = titleScale(pixel);
  const tops = titles.map((one) => one.y - titleBandHeight(one.title, one.count, one.width, scale) - titleAbove(scale));
  const top = Math.min(bounds.y, ...tops);
  return { ...bounds, y: top, height: bounds.height + (bounds.y - top) };
}
