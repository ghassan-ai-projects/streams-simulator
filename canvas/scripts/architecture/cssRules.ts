import { readFileSync } from "node:fs";
import type { Register, Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { ImportRef, SourceFile } from "./sources.ts";
import { importsOf, moduleOf, walk } from "./sources.ts";

interface Match {
  readonly line: number;
  readonly text: string;
}

function matches(text: string, pattern: RegExp): Match[] {
  const out: Match[] = [];
  const lines = text.split("\n");
  lines.forEach((content, index) => {
    const found = content.match(pattern);
    if (found === null) return;
    for (const entry of found) out.push({ line: index + 1, text: entry });
  });
  return out;
}

function lineCount(text: string): number {
  return text.replace(/\n$/, "").split("\n").length;
}

function cssFiles(root: string): string[] {
  return walk(root).filter((path) => path.endsWith(".css"));
}

export function capRules(root: string, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const path of cssFiles(root)) {
    const actual = lineCount(readFileSync(path, "utf8"));
    const cap = register.css.caps[path]?.cap ?? register.css.defaultCap;
    if (actual > cap) {
      out.push(violation("css-cap", path, 1, `${String(actual)} lines exceeds the cap of ${String(cap)}`));
    }
  }
  return out;
}

export function globalLocationRules(root: string, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const path of cssFiles(root)) {
    const text = readFileSync(path, "utf8");
    const cap = register.css.globalCaps[path] ?? { global: 0, important: 0 };
    out.push(...overCap(":global", matches(text, /:global/g), cap.global, path));
    out.push(...overCap("!important", matches(text, /!important/g), cap.important, path));
  }
  return out;
}

function overCap(kind: string, found: readonly Match[], cap: number, path: string): Violation[] {
  return found
    .slice(cap)
    .map((entry) =>
      violation("css-global-location", path, entry.line, `${kind} exceeds the registered allowance of ${String(cap)}`),
    );
}

export function colorRules(root: string, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const path of cssFiles(root)) {
    if (register.css.globalEntrypoints.includes(path)) continue;
    const found = matches(readFileSync(path, "utf8"), /#[0-9a-fA-F]{3,8}\b|\brgba?\(|\bhsla?\(/g);
    const allowed = register.css.colorExceptions[path]?.literals ?? 0;
    out.push(
      ...found
        .slice(allowed)
        .map((entry) =>
          violation(
            "css-color-literal",
            path,
            entry.line,
            `${entry.text} is a raw colour; it belongs in the theme tokens`,
          ),
        ),
    );
  }
  return out;
}

export function ownershipRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    for (const imported of importsOf(file, register)) {
      if (!imported.target?.endsWith(".css")) continue;
      out.push(...ownershipViolation(file, imported, register));
    }
  }
  return out;
}

function ownershipViolation(file: SourceFile, imported: ImportRef, register: Register): Violation[] {
  const target = imported.target ?? "";
  if (register.css.globalEntrypoints.includes(target)) return [];
  const owner = moduleOf(file.path, register);
  const stylesheet = moduleOf(target, register);
  if (owner === stylesheet) return [];
  return [
    violation(
      "css-feature-ownership",
      target,
      imported.line,
      `${file.path} (${owner ?? "?"}) imports ${stylesheet ?? "?"} styles`,
      file.path,
    ),
  ];
}

export function composesRules(root: string, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const path of cssFiles(root)) {
    const owner = moduleOf(path, register);
    const found = matches(readFileSync(path, "utf8"), /composes:[^;]*from\s+"([^"]+)"/g);
    for (const entry of found) {
      const specifier = /from\s+"([^"]+)"/.exec(entry.text)?.[1];
      if (specifier === undefined) continue;
      const target = resolveCss(path, specifier);
      if (target !== null && moduleOf(target, register) === owner) continue;
      out.push(
        violation("css-composes-cross-module", path, entry.line, `composes crosses into ${target ?? specifier}`),
      );
    }
  }
  return out;
}

function resolveCss(from: string, specifier: string): string | null {
  const base = from.split("/").slice(0, -1);
  for (const part of specifier.split("/")) {
    if (part === ".") continue;
    if (part === "..") base.pop();
    else base.push(part);
  }
  const path = base.join("/");
  return path.endsWith(".css") ? path : `${path}.css`;
}
