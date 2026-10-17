import { readFileSync } from "node:fs";
import { parseJson, show } from "./git.ts";

interface GateFile {
  readonly path: string;
  readonly compare: (before: string, after: string) => string[];
}

const gateFiles: readonly GateFile[] = [
  { path: "vitest.config.ts", compare: coverageCompare },
  { path: "eslint.config.js", compare: eslintCompare },
  { path: ".jscpd.json", compare: cloneCompare },
  { path: "knip.json", compare: unusedCompare },
  { path: ".dependency-cruiser.cjs", compare: layerCompare },
  { path: ".prettierignore", compare: ignoreFileCompare },
  { path: "package.json", compare: scriptCompare },
];

export function gateRegressions(base: string, canvasDir: string): string[] {
  const out: string[] = [];
  for (const file of gateFiles) {
    const before = show(base, `${canvasDir}/${file.path}`);
    if (before !== null) out.push(...file.compare(before, readFileSync(file.path, "utf8")));
  }
  return out;
}

function numeric(
  what: string,
  before: ReadonlyMap<string, number>,
  after: ReadonlyMap<string, number>,
  worse: (was: number, now: number) => boolean,
): string[] {
  const out: string[] = [];
  for (const [name, value] of before) {
    const now = after.get(name);
    if (now === undefined) out.push(`${what} ${name} is gone`);
    else if (worse(value, now)) out.push(`${what} ${name} went from ${String(value)} to ${String(now)}`);
  }
  return out;
}

function grew(what: string, before: readonly string[], after: readonly string[]): string[] {
  const out: string[] = [];
  for (const entry of after) if (!before.includes(entry)) out.push(`${what} ${entry}`);
  return out;
}

function coverageCompare(before: string, after: string): string[] {
  return numeric("coverage floor", coverageThresholds(before), coverageThresholds(after), (was, now) => now < was);
}

function coverageThresholds(text: string): Map<string, number> {
  const out = new Map<string, number>();
  for (const match of text.matchAll(/"(src\/[^"]+)"\s*:\s*\{([^}]*)\}/g)) {
    for (const metric of (match[2] ?? "").matchAll(/(\w+):\s*(\d+)/g)) {
      out.set(`${match[1] ?? ""} ${metric[1] ?? ""}`, Number(metric[2]));
    }
  }
  return out;
}

function eslintCompare(before: string, after: string): string[] {
  return [
    ...numeric("lint cap", lintCaps(before), lintCaps(after), (was, now) => now > was),
    ...grew("eslint ignore gained", ignores(before), ignores(after)),
  ];
}

function lintCaps(text: string): Map<string, number> {
  const out = new Map<string, number>();
  for (const match of text.matchAll(/"(max-lines|max-lines-per-function)":\s*\["error",\s*\{\s*max:\s*(\d+)/g)) {
    out.set(match[1] ?? "", Number(match[2]));
  }
  for (const match of text.matchAll(/(complexity|sonarjs\/cognitive-complexity):\s*\["error",\s*(\d+)\]/g)) {
    out.set(match[1] ?? "", Number(match[2]));
  }
  return out;
}

function ignores(text: string): readonly string[] {
  const out: string[] = [];
  for (const block of text.matchAll(/ignores:\s*\[([^\]]*)\]/g)) {
    for (const entry of (block[1] ?? "").matchAll(/"([^"]+)"/g)) out.push(entry[1] ?? "");
  }
  return out;
}

function ignoreFileCompare(before: string, after: string): string[] {
  return grew("ignore entry", lines(before), lines(after));
}

function lines(text: string): readonly string[] {
  return text
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "");
}

function cloneCompare(before: string, after: string): string[] {
  const was = parseJson(before);
  const now = parseJson(after);
  const out: string[] = [];
  if (asNumber(now.minTokens) < asNumber(was.minTokens)) out.push("jscpd minTokens fell");
  if (asNumber(now.threshold) > asNumber(was.threshold)) out.push("jscpd threshold rose");
  if (asList(now.ignore).length > asList(was.ignore).length) out.push("jscpd ignore list grew");
  if (asNumber(now.exitCode) !== 1) out.push("jscpd no longer fails on a clone");
  return out;
}

function unusedCompare(before: string, after: string): string[] {
  const was = parseJson(before);
  const now = parseJson(after);
  const out: string[] = [];
  for (const key of ["entry", "project"]) {
    for (const entry of asList(was[key])) {
      if (!asList(now[key]).includes(entry)) out.push(`knip ${key} lost ${entry}`);
    }
  }
  for (const key of ["ignore", "ignoreDependencies", "ignoreBinaries"]) {
    if (asList(now[key]).length > asList(was[key]).length) out.push(`knip ${key} grew`);
  }
  return out;
}

function layerCompare(before: string, after: string): string[] {
  const out: string[] = [];
  const now = layerRules(after);
  for (const [name, severity] of layerRules(before)) {
    const current = now.get(name);
    if (current === undefined) out.push(`dependency-cruiser rule ${name} is gone`);
    else if (current !== severity) out.push(`dependency-cruiser rule ${name} severity changed`);
  }
  for (const exclusion of exclusions(after)) {
    if (!exclusions(before).includes(exclusion)) out.push(`dependency-cruiser gained exclusion ${exclusion}`);
  }
  return out;
}

function layerRules(text: string): Map<string, string> {
  const out = new Map<string, string>();
  for (const match of text.matchAll(/name:\s*"([^"]+)"[\s\S]{0,200}?severity:\s*"([^"]+)"/g)) {
    out.set(match[1] ?? "", match[2] ?? "");
  }
  return out;
}

function exclusions(text: string): readonly string[] {
  const out: string[] = [];
  for (const match of text.matchAll(/pathNot:\s*\[([^\]]*)\]/g)) {
    for (const entry of (match[1] ?? "").matchAll(/"([^"]+)"/g)) out.push(`pathNot:${entry[1] ?? ""}`);
  }
  for (const match of text.matchAll(/exclude:\s*\{\s*path:\s*"([^"]+)"/g)) out.push(`exclude:${match[1] ?? ""}`);
  return out;
}

function scriptCompare(before: string, after: string): string[] {
  const was = asRecord(parseJson(before).scripts);
  const now = asRecord(parseJson(after).scripts);
  const out: string[] = [];
  for (const [name, command] of Object.entries(was)) {
    const current = now[name];
    if (typeof current !== "string") out.push(`package.json script ${name} is gone`);
    else {
      out.push(...flagRegressions(name, String(command), current));
      out.push(...operatorRegressions(name, String(command), current));
    }
  }
  for (const command of subCommands(String(was.check))) {
    if (!String(now.check).includes(command)) out.push(`the check chain lost ${command}`);
  }
  return out;
}

function flagRegressions(script: string, before: string, after: string): string[] {
  const out: string[] = [];
  const tokens = before.split(" ").filter((token) => token !== "");
  tokens.forEach((token, index) => {
    if (!token.startsWith("--")) return;
    const [name = token, inline] = token.split("=");
    const next = tokens[index + 1] ?? "";
    const limit = inline === undefined ? (/^\d+$/.test(next) ? Number(next) : Number.NaN) : Number(inline);
    if (Number.isNaN(limit)) {
      if (!after.includes(name)) out.push(`package.json script ${script} lost ${token}`);
      return;
    }
    const found = new RegExp(`${name}[= ](\\d+)`).exec(after);
    if (found === null) out.push(`package.json script ${script} lost ${token}`);
    else if (Number(found[1]) > limit) out.push(`package.json script ${script} loosened ${token}`);
  });
  return out;
}

function operatorRegressions(script: string, before: string, after: string): string[] {
  const out: string[] = [];
  for (const operator of ["||", ";", " & ", "|"]) {
    if (!before.includes(operator) && after.includes(operator)) {
      out.push(`package.json script ${script} gained the shell operator ${operator.trim()}`);
    }
  }
  return out;
}

function subCommands(chain: string): readonly string[] {
  return chain.split("&&").map((part) => part.trim());
}

function asNumber(value: unknown): number {
  return typeof value === "number" ? value : 0;
}

function asList(value: unknown): readonly string[] {
  return Array.isArray(value) ? value.map(String) : [];
}

function asRecord(value: unknown): Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}
