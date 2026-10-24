export function endText(name: string, atS: number, untilS: number | undefined, durationS: number): string | null {
  if (untilS === undefined) return null;
  if (!Number.isFinite(untilS)) return `${name} ends at a number of minutes.`;
  if (Number.isFinite(atS) && untilS <= atS) return `${name} ends after it starts.`;
  return untilS > durationS ? `${name} ends by the end of the run.` : null;
}
