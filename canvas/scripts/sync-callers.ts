import { readFileSync, writeFileSync } from "node:fs";
import { readRegister } from "./architecture/register.ts";
import { importsOf, readSources } from "./architecture/sources.ts";

/**
 * Rewrites the register's `callers` fields from the observed import graph, so the recorded callers cannot drift from
 * the tree. `node scripts/sync-callers.ts [--check]` reports instead of writing when `--check` is given.
 */

const registerPath = "architecture.json";

export function observedCallers(): Map<string, string[]> {
  const register = readRegister(registerPath);
  const graph = new Map<string, Set<string>>();
  for (const name of Object.keys(register.modules)) graph.set(name, new Set());
  for (const file of readSources(register)) {
    for (const imported of importsOf(file, register)) {
      if (imported.module === null || imported.module === file.module) continue;
      graph.get(imported.module)?.add(file.module);
    }
  }
  return new Map([...graph].map(([name, callers]) => [name, [...callers].sort()]));
}

interface RegisterFile {
  readonly modules: Record<string, { callers?: readonly string[] }>;
}

function readRegisterFile(): RegisterFile {
  const parsed: unknown = JSON.parse(readFileSync(registerPath, "utf8"));
  if (typeof parsed !== "object" || parsed === null) throw new Error(`${registerPath} must hold an object`);
  return parsed as RegisterFile;
}

function applyObserved(file: RegisterFile, observed: ReadonlyMap<string, string[]>): number {
  let changed = 0;
  for (const [name, callers] of observed) {
    const spec = file.modules[name];
    if (spec === undefined || [...(spec.callers ?? [])].sort().join(",") === callers.join(",")) continue;
    changed += 1;
    process.stdout.write(
      `${name}: ${(spec.callers ?? []).join(", ") || "(none)"} -> ${callers.join(", ") || "(none)"}\n`,
    );
    spec.callers = callers;
  }
  return changed;
}

function main(): void {
  const file = readRegisterFile();
  const changed = applyObserved(file, observedCallers());
  if (process.argv.includes("--check") || changed === 0) {
    process.stdout.write(`canvas callers: ${changed === 0 ? "ok" : `${String(changed)} out of date`}\n`);
    process.exitCode = changed === 0 ? 0 : 1;
    return;
  }
  writeFileSync(registerPath, `${JSON.stringify(file, null, 2)}\n`);
  process.stdout.write(`canvas callers: ${String(changed)} updated in ${registerPath}\n`);
}

main();
