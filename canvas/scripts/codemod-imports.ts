import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, relative, resolve } from "node:path";
import ts from "typescript";
import { walk } from "./architecture/sources.ts";

/**
 * Import codemod for module moves: reads a move map, rewrites every relative import and re-export that points at a
 * moved file, and reports the edits. `--apply` writes them; without it nothing changes.
 *
 * node scripts/codemod-imports.ts moves.json [--apply]
 * moves.json: { "src/old/File.ts": "src/new/internal/File.ts", ... }
 */

const extensions = [".ts", ".tsx", ".css"];

export interface Edit {
  readonly file: string;
  readonly from: string;
  readonly to: string;
  readonly line: number;
}

/** The path a specifier names, whether it still exists (an unmoved file) or only existed before the move. */
export function resolveTarget(
  fromPath: string,
  specifier: string,
  moves: Readonly<Record<string, string>> = {},
): string | null {
  if (!specifier.startsWith(".")) return null;
  const base = resolve("/", fromPath, "..", specifier).slice(1);
  const candidates = [base, ...extensions.map((ext) => base + ext)];
  return (
    candidates.find((candidate) => moves[candidate] !== undefined || (existsSync(candidate) && isFile(candidate))) ??
    null
  );
}

function isFile(path: string): boolean {
  try {
    return readFileSync(path, "utf8") !== null;
  } catch {
    return false;
  }
}

function specifierFor(fromPath: string, targetPath: string): string {
  const bare = targetPath.replace(/\.(tsx?|css)$/, "");
  const relativePath = relative(dirname(fromPath), bare).replaceAll("\\", "/");
  return relativePath.startsWith(".") ? relativePath : `./${relativePath}`;
}

export interface MovePlan {
  readonly moves: Readonly<Record<string, string>>;
  /** module folder -> its entrypoint, for imports that come from outside the module */
  readonly facades: Readonly<Record<string, string>>;
}

function entrypointFor(target: string, importer: string, facades: Readonly<Record<string, string>>): string | null {
  for (const [folder, entry] of Object.entries(facades)) {
    if (target === entry || target.startsWith(`${folder}/`)) {
      return importer === entry || importer.startsWith(`${folder}/`) ? null : entry;
    }
  }
  return null;
}

export function planFile(path: string, plan: MovePlan): Edit[] {
  const text = readFileSync(path, "utf8");
  const source = ts.createSourceFile(path, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  const edits: Edit[] = [];
  const visit = (node: ts.Node): void => {
    if (
      (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) &&
      node.moduleSpecifier !== undefined &&
      ts.isStringLiteral(node.moduleSpecifier)
    ) {
      const specifier = node.moduleSpecifier.text;
      const target = resolveTarget(path, specifier, plan.moves);
      if (target === null) return;
      const moved = plan.moves[target] ?? target;
      const entry = entrypointFor(moved, path, plan.facades);
      const destination = entry ?? moved;
      if (destination === target) return;
      edits.push({
        file: path,
        from: specifier,
        to: specifierFor(path, destination),
        line: source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1,
      });
    }
    node.forEachChild(visit);
  };
  source.forEachChild(visit);
  return edits;
}

export function planAll(root: string, plan: MovePlan): Edit[] {
  return walk(root)
    .filter((path) => /\.tsx?$/.test(path))
    .flatMap((path) => planFile(path, plan));
}

export function applyEdits(edits: readonly Edit[]): void {
  const byFile = new Map<string, Edit[]>();
  for (const edit of edits) byFile.set(edit.file, [...(byFile.get(edit.file) ?? []), edit]);
  for (const [file, list] of byFile) {
    let text = readFileSync(file, "utf8");
    for (const edit of list) {
      text = text.replace(new RegExp(`(from\\s+|import\\s*)["']${escape(edit.from)}["']`, "g"), `$1"${edit.to}"`);
    }
    writeFileSync(file, text);
  }
}

function escape(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function planFrom(path: string): MovePlan {
  const value: unknown = JSON.parse(readFileSync(path, "utf8"));
  if (typeof value !== "object" || value === null || Array.isArray(value))
    throw new Error(`${path} must hold an object`);
  const source = value as { moves?: Record<string, string>; facades?: Record<string, string> };
  if (source.moves === undefined && !Array.isArray(value)) {
    const legacy = value as Record<string, string>;
    return { moves: legacy, facades: {} };
  }
  return { moves: source.moves ?? {}, facades: source.facades ?? {} };
}

function main(): void {
  const [mapPath = "", flag] = process.argv.slice(2);
  if (mapPath === "") throw new Error("usage: node scripts/codemod-imports.ts moves.json [--apply]");
  const plan = planFrom(mapPath);
  const edits = planAll("src", plan);
  for (const edit of edits) {
    process.stdout.write(`${edit.file}:${String(edit.line)} ${edit.from} -> ${edit.to}\n`);
  }
  process.stdout.write(
    `${String(edits.length)} imports to rewrite in ${String(new Set(edits.map((e) => e.file)).size)} files\n`,
  );
  if (flag === "--apply") {
    applyEdits(edits);
    process.stdout.write("applied\n");
  }
}

if (process.argv[1]?.endsWith("codemod-imports.ts") === true) main();
