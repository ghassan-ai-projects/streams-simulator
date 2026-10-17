import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { checkRegister } from "./check.ts";
import { mergeRegister } from "./mergeRegister.ts";
import { runMergeFixtures } from "./mergeFixtures.ts";
import { parseRegister } from "./register.ts";
import { applyDebt, describe, sortViolations } from "./violations.ts";
import type { Register } from "./model.ts";
import { runRatchetFixtures } from "./ratchetFixtures.ts";

const defaultBase = "base.register.json";
const patchFile = "register.patch.json";

interface Expected {
  readonly base: string;
  readonly rules: readonly string[];
  readonly stale: readonly string[];
}

export function runFixtures(root: string): string[] {
  const failures: string[] = [];
  const cwd = process.cwd();
  const names = readdirSync(root, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort();
  for (const name of names) {
    failures.push(...runFixture(root, name, cwd));
  }
  failures.push(...runMergeFixtures());
  failures.push(...runRatchetFixtures());
  const verdict = failures.length === 0 ? "ok" : `FAILED (${String(failures.length)})`;
  process.stdout.write(`canvas fixtures: ${verdict}, ${String(names.length)} cases\n`);
  return failures;
}

function runFixture(root: string, name: string, cwd: string): string[] {
  const dir = join(cwd, root, name);
  const expected = readExpected(dir);
  const register = buildRegister(join(cwd, root), dir, expected.base);
  process.chdir(dir);
  const result = checkRegister(register, "all");
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

function buildRegister(root: string, dir: string, base: string): Register {
  const patchPath = join(dir, patchFile);
  const merged = mergeRegister(readJson(join(root, base)), existsSync(patchPath) ? readJson(patchPath) : {});
  return parseRegister(merged, `${dir} (${base})`);
}

function readJson(path: string): unknown {
  return JSON.parse(readFileSync(path, "utf8"));
}

function readExpected(dir: string): Expected {
  const value = readJson(join(dir, "expected.json"));
  const source = typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
  const rules = Array.isArray(source.rules) ? source.rules.map(String) : [];
  const stale = Array.isArray(source.stale) ? source.stale.map(String) : [];
  const base = typeof source.base === "string" ? source.base : defaultBase;
  return { base, rules: [...rules].sort(), stale: [...stale].sort() };
}
