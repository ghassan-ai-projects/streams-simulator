import { emittedCssRules } from "./architecture/emitted.ts";
import { readRegister } from "./architecture/register.ts";
import { describe } from "./architecture/violations.ts";

const register = readRegister("architecture.json");
const violations = emittedCssRules("dist", register.emittedCss);
for (const entry of violations) process.stdout.write(`  ${describe(entry)}\n`);
process.stdout.write(
  `canvas emitted css: ${violations.length === 0 ? "ok" : `FAILED (${String(violations.length)})`}\n`,
);
process.exitCode = violations.length === 0 ? 0 : 1;
