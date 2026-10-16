import { readFileSync } from "node:fs";
import { appClockRules, appUnknownRules, booleanParameterRules, testTimerRules } from "./codeRules.ts";
import { capRules, colorRules, composesRules, globalLocationRules, ownershipRules } from "./cssRules.ts";
import type { Register, Violation } from "./model.ts";
import { readRegister } from "./register.ts";
import { readSources } from "./sources.ts";
import { entrypointRules, inlineStyleRules } from "./styleRules.ts";
import {
  callerRules,
  doubleAssertionRules,
  moduleRules,
  protocolRules,
  restrictedRules,
  targetRules,
} from "./tsRules.ts";

export type RuleSet = "architecture" | "css" | "all";

const cssRuleNames = [
  "css-cap",
  "css-global-location",
  "css-color-literal",
  "css-feature-ownership",
  "css-composes-cross-module",
];

export interface CheckResult {
  readonly register: Register;
  readonly violations: readonly Violation[];
}

export function check(registerPath: string, ruleSet: RuleSet): CheckResult {
  const register = readRegister(registerPath);
  const files = readSources(register);
  const violations = [
    ...moduleRules(files, register),
    ...targetRules(files, register),
    ...callerRules(files, register),
    ...protocolRules(files, register),
    ...restrictedRules(files, register),
    ...doubleAssertionRules(files, register),
    ...appUnknownRules(files),
    ...appClockRules(files, register),
    ...inlineStyleRules(files, register),
    ...entrypointRules(files, register),
    ...booleanParameterRules(files, register),
    ...testTimerRules(files, register),
    ...capRules(register.source, register),
    ...globalLocationRules(register.source, register),
    ...colorRules(register.source, register),
    ...ownershipRules(files, register),
    ...composesRules(register.source, register),
  ];
  return { register, violations: violations.filter((entry) => wanted(entry, ruleSet)) };
}

function wanted(entry: Violation, ruleSet: RuleSet): boolean {
  if (ruleSet === "all") return true;
  const isCss = cssRuleNames.includes(entry.rule);
  return ruleSet === "css" ? isCss : !isCss;
}

export function readText(path: string): string {
  return readFileSync(path, "utf8");
}
