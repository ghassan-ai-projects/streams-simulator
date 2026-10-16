import type { CheckResult } from "./check.ts";
import { applyDebt, describe, sortViolations } from "./violations.ts";

export function report(name: string, result: CheckResult, ruleSet: "architecture" | "css"): number {
  const scoped = result.register.debt.filter((entry) => entry.rule.startsWith("css-") === (ruleSet === "css"));
  const { remaining, stale, exceeded, suppressed } = applyDebt(result.violations, scoped);
  for (const entry of sortViolations(remaining)) process.stdout.write(`  ${describe(entry)}\n`);
  for (const entry of exceeded) process.stdout.write(`  debt exceeded: ${entry}\n`);
  for (const entry of stale) process.stdout.write(`  stale debt: ${entry}\n`);
  const problems = remaining.length + stale.length + exceeded.length;
  const verdict = problems === 0 ? "ok" : `FAILED (${String(problems)})`;
  process.stdout.write(`${name}: ${verdict}, ${String(suppressed)} registered debt\n`);
  return problems === 0 ? 0 : 1;
}
