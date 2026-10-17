import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import type { Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { SourceFile } from "./sources.ts";
import { resolveSpecifier, walk } from "./sources.ts";

interface Composition {
  readonly sheet: string;
  readonly names: readonly string[];
}

interface Sheet {
  readonly path: string;
  readonly defined: ReadonlySet<string>;
  readonly compositions: readonly Composition[];
}

interface Use {
  readonly file: string;
  readonly line: number;
  readonly sheet: string;
  readonly keys: ReadonlySet<string>;
  readonly dynamic: boolean;
}

export function cssKeyRules(files: readonly SourceFile[], root: string): Violation[] {
  const sheets = new Map(
    walk(root)
      .filter((path) => path.endsWith(".module.css"))
      .map((path) => [path, readSheet(path)]),
  );
  const facades = facadeSheets(files);
  const uses = files.flatMap((file) => usesIn(file, sheets, facades));
  return [...undefinedKeys(uses, sheets), ...unusedKeys(uses, sheets)];
}

function undefinedKeys(uses: readonly Use[], sheets: ReadonlyMap<string, Sheet>): Violation[] {
  return uses.flatMap((use) =>
    [...use.keys]
      .filter((key) => !(sheets.get(use.sheet)?.defined.has(key) ?? true))
      .map((key) =>
        violation("css-class-undefined", use.file, use.line, `${key} is not a class of ${use.sheet}`, use.sheet),
      ),
  );
}

function unusedKeys(uses: readonly Use[], sheets: ReadonlyMap<string, Sheet>): Violation[] {
  const out: Violation[] = [];
  for (const sheet of sheets.values()) {
    const users = uses.filter((use) => use.sheet === sheet.path);
    if (users.length === 0 || users.some((use) => use.dynamic)) continue;
    const composed = [...sheets.values()]
      .flatMap((other) => other.compositions)
      .filter((entry) => entry.sheet === sheet.path);
    const used = new Set([...users.flatMap((use) => [...use.keys]), ...composed.flatMap((entry) => entry.names)]);
    for (const key of sheet.defined) {
      if (!used.has(key)) out.push(violation("css-class-unused", sheet.path, 1, `.${key} is defined and never used`));
    }
  }
  return out;
}

function readSheet(path: string): Sheet {
  const text = readFileSync(path, "utf8").replace(/\/\*[\s\S]*?\*\//g, "");
  const defined = new Set<string>();
  for (const match of text.matchAll(/([^{};]*)\{/g)) {
    const prelude = (match[1] ?? "").trim();
    if (prelude.startsWith("@")) continue;
    const plain = prelude
      .replace(/:global\([^)]*\)/g, "")
      .replace(/\[[^\]]*\]/g, "")
      .replace(/\([^)]*\)/g, "");
    for (const name of plain.matchAll(/\.([A-Za-z_][\w-]*)/g)) defined.add(name[1] ?? "");
  }
  const compositions = [...text.matchAll(/composes:\s*([\w\s-]+?)\s*(?:from\s*["']([^"']+)["'])?\s*;/g)].map(
    (match) => ({
      sheet: match[2] === undefined ? path : join(dirname(path), match[2]),
      names: (match[1] ?? "").split(/\s+/),
    }),
  );
  return { path, defined, compositions };
}

function facadeSheets(files: readonly SourceFile[]): Map<string, Map<string, string>> {
  const out = new Map<string, Map<string, string>>();
  for (const file of files.filter((entry) => entry.path.endsWith("/index.ts"))) {
    const names = new Map<string, string>();
    const text = file.source.getFullText();
    for (const match of text.matchAll(/export \{ default as (\w+) \} from "([^"]+\.module\.css)"/g)) {
      names.set(match[1] ?? "", join(dirname(file.path), match[2] ?? ""));
    }
    if (names.size > 0) out.set(file.path, names);
  }
  return out;
}

function usesIn(
  file: SourceFile,
  sheets: ReadonlyMap<string, Sheet>,
  facades: ReadonlyMap<string, ReadonlyMap<string, string>>,
): Use[] {
  const text = file.source.getFullText();
  const body = text.replace(/^import [^;]*;$/gm, "");
  return [...bindingsOf(file, text, facades)]
    .filter(([, sheet]) => sheets.has(sheet))
    .map(([local, sheet]) => useOf(file, body, local, sheet));
}

function bindingsOf(
  file: SourceFile,
  text: string,
  facades: ReadonlyMap<string, ReadonlyMap<string, string>>,
): Map<string, string> {
  const bound = new Map<string, string>();
  for (const match of text.matchAll(/import (\w+) from "([^"]+\.module\.css)"/g)) {
    bound.set(match[1] ?? "", join(dirname(file.path), match[2] ?? ""));
  }
  for (const match of text.matchAll(/import \{([^}]*)\} from "([^"]+)"/g)) {
    const target = resolveSpecifier(file.path, match[2] ?? "");
    const names = target === null ? undefined : facades.get(target);
    if (names !== undefined) addFacadeBindings(bound, names, match[1] ?? "");
  }
  return bound;
}

function addFacadeBindings(bound: Map<string, string>, names: ReadonlyMap<string, string>, list: string): void {
  for (const part of list.split(",")) {
    const [imported = "", local = imported] = part.trim().split(/\s+as\s+/);
    const sheet = names.get(imported);
    if (sheet !== undefined) bound.set(local, sheet);
  }
}

function useOf(file: SourceFile, text: string, local: string, sheet: string): Use {
  const keys = new Set<string>();
  let line = 1;
  const pattern = new RegExp(`(?<![\\w.])${local}\\.([A-Za-z_][\\w-]*)|(?<![\\w.])${local}\\["([\\w-]+)"\\]`, "g");
  for (const match of text.matchAll(pattern)) {
    keys.add(match[1] ?? match[2] ?? "");
    if (line === 1) line = text.slice(0, match.index).split("\n").length;
  }
  const dynamic = new RegExp(`(?<![\\w.])${local}(?![\\w.]|\\["|\\[\\s*")`).test(text);
  return { file: file.path, line, sheet, keys, dynamic };
}
