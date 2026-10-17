import type { DebtEntry, Violation } from "./model.ts";
import { debtKey, violationKey } from "./model.ts";

export interface DebtReport {
  readonly remaining: readonly Violation[];
  readonly stale: readonly string[];
  readonly exceeded: readonly string[];
  readonly suppressed: number;
}

export function applyDebt(violations: readonly Violation[], debt: readonly DebtEntry[]): DebtReport {
  const grouped = group(violations);
  const entries = new Map(debt.map((entry) => [debtKey(entry), entry]));
  const remaining: Violation[] = [];
  const stale: string[] = [];
  const exceeded: string[] = [];
  let suppressed = 0;
  for (const [key, found] of grouped) {
    const entry = entries.get(key);
    if (entry === undefined) {
      remaining.push(...found);
      continue;
    }
    if (found.length > entry.count) {
      remaining.push(...found.slice(entry.count));
      exceeded.push(`${subject(entry)}: ${String(found.length)} found, ${String(entry.count)} registered`);
    }
    if (found.length < entry.count) {
      stale.push(`${subject(entry)}: now ${String(found.length)}, registered ${String(entry.count)}`);
    }
    suppressed += Math.min(found.length, entry.count);
  }
  for (const [key, entry] of entries) {
    if (!grouped.has(key)) stale.push(`${subject(entry)}: gone, remove the entry (${entry.phase})`);
  }
  return { remaining, stale, exceeded, suppressed };
}

function group(violations: readonly Violation[]): Map<string, Violation[]> {
  const out = new Map<string, Violation[]>();
  for (const entry of violations) {
    const key = violationKey(entry);
    out.set(key, [...(out.get(key) ?? []), entry]);
  }
  return out;
}

function subject(entry: { readonly rule: string; readonly file: string; readonly target?: string }): string {
  return [entry.rule, entry.file, entry.target ?? ""].filter((part) => part !== "").join(" ");
}

export function describe(entry: Violation): string {
  const where = entry.target === undefined ? entry.file : `${entry.file} <- ${entry.target}`;
  return `${entry.rule} ${where}:${String(entry.line)} ${entry.message}`;
}

export function sortViolations(violations: readonly Violation[]): Violation[] {
  return [...violations].sort((a, b) =>
    `${a.rule}${a.file}${String(a.line)}`.localeCompare(`${b.rule}${b.file}${String(b.line)}`),
  );
}
