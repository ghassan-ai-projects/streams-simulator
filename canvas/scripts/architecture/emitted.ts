import { readFileSync } from "node:fs";
import { globSync } from "node:fs";
import type { Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { EmittedSpec } from "./model.ts";

export function emittedCssRules(distDir: string, spec: EmittedSpec): Violation[] {
  const files = globSync(`${distDir}/assets/index-*.css`).sort();
  const name = files[0];
  if (name === undefined) {
    return [violation("emitted-css", `${distDir}/assets`, 1, "no built stylesheet; run npm run build first")];
  }
  return [...orderRules(name, spec), ...declarationRules(name, spec)];
}

function orderRules(path: string, spec: EmittedSpec): Violation[] {
  const text = readFileSync(path, "utf8");
  const out: Violation[] = [];
  let previous = -1;
  let previousName = "";
  for (const entry of spec.order) {
    const at = text.indexOf(entry.selector);
    if (at < 0) {
      out.push(
        violation(
          "emitted-css-order",
          path,
          1,
          `${entry.name} (${entry.selector}) is missing from the built stylesheet`,
        ),
      );
      continue;
    }
    if (at < previous) {
      out.push(
        violation(
          "emitted-css-order",
          path,
          1,
          `${entry.name} now precedes ${previousName}; the cascade order changed`,
        ),
      );
    }
    previous = at;
    previousName = entry.name;
  }
  return out;
}

function declarationRules(path: string, spec: EmittedSpec): Violation[] {
  const text = readFileSync(path, "utf8");
  const out: Violation[] = [];
  out.push(...countRule(path, "!important", count(text, /!important/g), spec.important));
  out.push(...countRule(path, "z-index", count(text, /z-index\s*:/g), spec.zIndex));
  out.push(...countRule(path, "positioned", count(text, /position\s*:\s*(absolute|fixed|sticky)/g), spec.positioned));
  return out;
}

function countRule(path: string, what: string, actual: number, expected: number): Violation[] {
  if (actual === expected) return [];
  return [
    violation(
      "emitted-css-declarations",
      path,
      1,
      `${what} declarations changed ${String(expected)} -> ${String(actual)}; name the correction in the register`,
    ),
  ];
}

function count(text: string, pattern: RegExp): number {
  return [...text.matchAll(pattern)].length;
}
