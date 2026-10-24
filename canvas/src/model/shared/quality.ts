export type Quality = "" | "suspect" | "out_of_range" | "stale" | "disconnected";

const words: Readonly<Record<Quality, string>> = {
  "": "",
  suspect: "suspect",
  out_of_range: "out of range",
  stale: "stale",
  disconnected: "disconnected",
};

export function qualityWords(quality: Quality): string {
  return words[quality];
}
