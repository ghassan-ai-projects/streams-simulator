import { readFileSync } from "node:fs";
import { gateRegressions } from "./architecture/gateRatchet.ts";
import { repositoryRoot } from "./architecture/git.ts";
import { registerAt, registerBase, registerRegressions, shallowProblems } from "./architecture/ratchet.ts";
import type { Register } from "./architecture/model.ts";
import { readRegister } from "./architecture/register.ts";

const registerPath = "architecture.json";

/** The register's floors and the enforced thresholds must describe the same numbers. */
function floorDrift(register: Register, vitest: string): string[] {
  const out: string[] = [];
  for (const [glob, floors] of Object.entries(register.floors)) {
    const block = new RegExp(`"${glob.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}"\\s*:\\s*\\{([^}]*)\\}`).exec(vitest);
    if (block === null) {
      out.push(`coverage floor ${glob} is not in vitest.config.ts`);
      continue;
    }
    for (const [metric, value] of Object.entries(floors)) {
      const enforced = new RegExp(`${metric}:\\s*(\\d+)`).exec(block[1] ?? "");
      if (enforced === null) out.push(`coverage floor ${glob} ${metric} is not enforced in vitest.config.ts`);
      else if (Number(enforced[1]) < value)
        out.push(
          `coverage floor ${glob} ${metric} is enforced at ${enforced[1] ?? ""}, below the register's ${String(value)}`,
        );
    }
  }
  return out;
}
const canvasDir = "canvas";
const problems: string[] = [];

problems.push(...shallowProblems());
const base = problems.length > 0 ? null : registerBase(registerPath);
const previous = base === null ? null : registerAt(base.commit, `${canvasDir}/${registerPath}`);
if (problems.length === 0 && (base === null || previous === null)) {
  process.stdout.write("canvas ratchet: no baseline yet; the first commit of architecture.json becomes it\n");
} else if (base !== null && previous !== null) {
  const register = readRegister(registerPath);
  problems.push(...registerRegressions(previous, register));
  problems.push(...gateRegressions(base.commit, canvasDir));
  problems.push(...floorDrift(register, readFileSync("vitest.config.ts", "utf8")));
}

for (const problem of problems) process.stdout.write(`  ${problem}\n`);
process.stdout.write(
  `canvas ratchet: ${problems.length === 0 ? "ok" : `FAILED (${String(problems.length)})`} in ${repositoryRoot()}\n`,
);
process.exitCode = problems.length === 0 ? 0 : 1;
