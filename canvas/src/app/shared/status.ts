export interface SaveStatus {
  readonly text: string;
  readonly detail: string | null;
  readonly tone: "busy" | "warn" | "ok" | "";
}

export function status(text: string, tone: SaveStatus["tone"], detail: string | null = null): SaveStatus {
  return { text, detail, tone };
}
