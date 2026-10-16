import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { check } from "./check.ts";
import { applyDebt, describe, sortViolations } from "./violations.ts";

interface Expected {
  readonly rules: readonly string[];
  readonly stale: readonly string[];
}

export function runFixtures(root: string): string[] {
  const failures: string[] = [];
  const cwd = process.cwd();
  const names = readdirSync(root).sort();
  for (const name of names) {
    failures.push(...runFixture(root, name, cwd));
  }
  const verdict = failures.length === 0 ? "ok" : `FAILED (${String(failures.length)})`;
  process.stdout.write(`canvas fixtures: ${verdict}, ${String(names.length)} cases\n`);
  return failures;
}

function runFixture(root: string, name: string, cwd: string): string[] {
  const dir = join(cwd, root, name);
  const expected = readExpected(dir);
  process.chdir(dir);
  const result = check("register.json", "all");
  const { remaining, stale } = applyDebt(result.violations, result.register.debt);
  const actualRules = [...new Set(remaining.map((entry) => entry.rule))].sort();
  const actualStale = stale.map((entry) => entry.split(":")[0]?.split("\u0000").join(" ") ?? entry).sort();
  process.chdir(cwd);
  const out: string[] = [];
  if (expected.rules.join(",") !== actualRules.join(",")) {
    out.push(`${name}: expected rules [${expected.rules.join(", ")}], found [${actualRules.join(", ")}]`);
  }
  if (expected.stale.join(",") !== actualStale.join(",")) {
    out.push(`${name}: expected stale [${expected.stale.join(", ")}], found [${actualStale.join(", ")}]`);
  }
  if (remaining.length > 0 && expected.rules.length === 0) {
    out.push(`${name}: unexpected violations [${sortViolations(remaining).map(describe).join("; ")}]`);
  }
  return out;
}

function readExpected(dir: string): Expected {
  const value: unknown = JSON.parse(readFileSync(join(dir, "expected.json"), "utf8"));
  const source = typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
  const rules = Array.isArray(source.rules) ? source.rules.map(String) : [];
  const stale = Array.isArray(source.stale) ? source.stale.map(String) : [];
  return { rules: [...rules].sort(), stale: [...stale].sort() };
}
