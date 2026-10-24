import { counted, shortDigest } from "../../shared/words";

export interface PlausibilityWords {
  readonly state: "ok" | "warn" | "bad";
  readonly text: string;
}

export function plausibilityWords(errors: number, warnings: number): PlausibilityWords {
  if (errors > 0) return { state: "bad", text: counted(errors, "issue") };
  if (warnings > 0) return { state: "warn", text: `Plausible, with ${counted(warnings, "warning")}` };
  return { state: "ok", text: "Readings stay plausible in a nominal day" };
}

export function digestWords(digest: string): string {
  return digest === "" ? "not reported" : shortDigest(digest);
}
