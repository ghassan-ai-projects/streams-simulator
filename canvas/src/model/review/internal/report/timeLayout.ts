export interface ChipPlace {
  readonly left: number;
  readonly row: number;
}

export function placeChips(
  positions: readonly number[],
  widths: readonly number[],
  lane: { readonly width: number; readonly rows: number; readonly gap: number },
): readonly ChipPlace[] {
  const ends: number[] = [];
  return positions.map((position, index) => {
    const width = widths[index] ?? 0;
    const left = Math.max(0, Math.min(position - width / 2, lane.width - width));
    const free = ends.findIndex((end) => end <= left);
    const row = free >= 0 ? free : ends.length < lane.rows ? ends.length : -1;
    if (row >= 0) ends[row] = left + width + lane.gap;
    return { left, row };
  });
}

const tickSteps = [600e9, 900e9, 1800e9, 3600e9, 7200e9, 10_800e9, 21_600e9, 43_200e9, 86_400e9];

export function timeTicks(startNs: number, endNs: number, most = 9): readonly number[] {
  const span = endNs - startNs;
  if (span <= 0) return [];
  const step = tickSteps.find((candidate) => span / candidate <= most) ?? span;
  const first = Math.ceil(startNs / step) * step;
  return Array.from({ length: Math.floor((endNs - first) / step) + 1 }, (_, index) => first + index * step);
}
