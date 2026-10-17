import { mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { join, relative } from "node:path";
import ts from "typescript";

/**
 * Drafts a module facade from the import graph: `contract.ts` re-exports the public types, `index.ts` re-exports the
 * values and the contract. Reads the move map produced for Phase 5 and writes one pair per module that has callers.
 *
 * node scripts/draft-facades.ts p5-moves.json [--write]
 */

interface Needed {
  readonly values: Map<string, Set<string>>;
  readonly types: Map<string, Set<string>>;
}

const foundations = new Set(["model/shared", "app/ports", "app/shared", "test", "root"]);

function sources(root = "src"): string[] {
  const out: string[] = [];
  const walk = (dir: string): void => {
    for (const entry of readdirSync(dir)) {
      const path = join(dir, entry);
      if (statSync(path).isDirectory()) walk(path);
      else if (/\.tsx?$/.test(path)) out.push(path);
    }
  };
  walk(root);
  return out;
}

export function draft(registerPath: string, movesPath: string): Map<string, Needed> {
  const register: unknown = JSON.parse(readFileSync(registerPath, "utf8"));
  const moves: Record<string, string> = JSON.parse(readFileSync(movesPath, "utf8")) as Record<string, string>;
  const targets = (register as { targets: Record<string, string> }).targets;
  const files = sources();
  const after = (path: string): string => moves[path] ?? path;
  const targetOf = (path: string): string | null => {
    if (targets[path] !== undefined) return targets[path] ?? null;
    const keys = Object.keys(targets)
      .filter((key) => path.startsWith(`${key}/`))
      .sort((a, b) => b.length - a.length);
    return keys[0] === undefined ? null : (targets[keys[0]] ?? null);
  };
  const needed = new Map<string, Needed>();
  for (const file of files) {
    const source = ts.createSourceFile(
      file,
      readFileSync(file, "utf8"),
      ts.ScriptTarget.Latest,
      true,
      ts.ScriptKind.TSX,
    );
    for (const statement of source.statements) {
      if (!ts.isImportDeclaration(statement) || !ts.isStringLiteral(statement.moduleSpecifier)) continue;
      const specifier = statement.moduleSpecifier.text;
      if (!specifier.startsWith(".")) continue;
      const base = relative(".", join(file, "..", specifier)).replaceAll("\\", "/");
      const resolved = [base, `${base}.ts`, `${base}.tsx`].find((candidate) => files.includes(candidate));
      const clause = statement.importClause;
      if (resolved === undefined || clause === undefined) continue;
      const module = targetOf(resolved);
      if (module === null || module === targetOf(file) || foundations.has(module)) continue;
      const entry = needed.get(module) ?? { values: new Map(), types: new Map() };
      const record = (name: string, typeOnly: boolean): void => {
        const into = typeOnly ? entry.types : entry.values;
        const names = into.get(after(resolved)) ?? new Set<string>();
        names.add(name);
        into.set(after(resolved), names);
      };
      if (clause.name !== undefined) record(clause.name.text, clause.isTypeOnly);
      const bindings = clause.namedBindings;
      if (bindings !== undefined && ts.isNamedImports(bindings)) {
        for (const element of bindings.elements) record(element.name.text, clause.isTypeOnly || element.isTypeOnly);
      }
      needed.set(module, entry);
    }
  }
  return needed;
}

function facadeSource(module: string, needed: Needed): { readonly index: string; readonly contract: string } {
  const contract = [...needed.types.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([file, names]) => `export type { ${[...names].sort().join(", ")} } from "${internal(file, module)}";`);
  const values = [...needed.values.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([file, names]) => `export { ${[...names].sort().join(", ")} } from "${internal(file, module)}";`);
  const contractExports =
    contract.length === 0
      ? []
      : [
          `export type { ${[...needed.types.values()]
            .flatMap((names) => [...names])
            .sort()
            .join(", ")} } from "./contract";`,
        ];
  return { index: [...values, ...contractExports, ""].join("\n"), contract: [...contract, ""].join("\n") };
}

function internal(file: string, module: string): string {
  const rest = file.replace(new RegExp(`^src/${module}/`), "").replace(/\.(tsx?|css)$/, "");
  return `./${rest}`;
}

function main(): void {
  const [mapPath = "", flag] = process.argv.slice(2);
  if (mapPath === "") throw new Error("usage: node scripts/draft-facades.ts p5-moves.json [--write]");
  const needed = draft("architecture.json", mapPath);
  for (const [module, entry] of [...needed.entries()].sort()) {
    const { index, contract } = facadeSource(module, entry);
    process.stdout.write(
      `== ${module}: ${String(entry.values.size)} value files, ${String(entry.types.size)} type files\n`,
    );
    if (flag !== "--write") continue;
    mkdirSync(`src/${module}`, { recursive: true });
    writeFileSync(`src/${module}/contract.ts`, contract);
    writeFileSync(`src/${module}/index.ts`, index);
  }
  if (flag === "--write") process.stdout.write("facades written\n");
}

if (process.argv[1]?.endsWith("draft-facades.ts") === true) main();
