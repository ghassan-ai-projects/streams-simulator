import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, resolve } from "node:path";
import ts from "typescript";
import type { Register } from "./model.ts";

export interface SourceFile {
  readonly path: string;
  readonly text: string;
  readonly source: ts.SourceFile;
  readonly module: string;
  readonly isTest: boolean;
}

export interface ImportRef {
  readonly specifier: string;
  readonly typeOnly: boolean;
  readonly line: number;
  readonly target: string | null;
  readonly module: string | null;
}

const extensions = [".ts", ".tsx", ".css", ".js"];

export function walk(root: string, dir = root): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(dir).sort()) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) out.push(...walk(root, path));
    else out.push(path);
  }
  return out;
}

export function isTestFile(path: string): boolean {
  return /\.test\.tsx?$/.test(path) || path.startsWith("src/test/");
}

export function isProduction(path: string): boolean {
  return !isTestFile(path);
}

export function moduleOf(path: string, register: Register): string | null {
  const matches = Object.entries(register.modules).filter(
    ([, spec]) =>
      (spec.files ?? []).includes(path) ||
      (spec.folders ?? []).some((folder) => path === folder || path.startsWith(`${folder}/`)),
  );
  if (matches.length === 0) return null;
  return matches.sort((a, b) => longest(b[1]) - longest(a[1]))[0]?.[0] ?? null;
}

function longest(spec: { readonly folders?: readonly string[]; readonly files?: readonly string[] }): number {
  const lengths = [...(spec.folders ?? []), ...(spec.files ?? [])].map((entry) => entry.length);
  return Math.max(0, ...lengths);
}

export function resolveSpecifier(fromPath: string, specifier: string): string | null {
  if (!specifier.startsWith(".")) return null;
  const base = resolve("/", fromPath, "..", specifier).slice(1);
  const candidates = [base, ...extensions.map((ext) => base + ext), ...extensions.map((ext) => `${base}/index${ext}`)];
  return candidates.find((candidate) => existsSync(candidate) && isFile(candidate)) ?? null;
}

function isFile(path: string): boolean {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
}

export function readSources(register: Register): SourceFile[] {
  return walk(register.source)
    .filter((path) => /\.tsx?$/.test(path))
    .map((path) => parseSource(path, readFileSync(path, "utf8"), register));
}

export function parseSource(path: string, text: string, register: Register): SourceFile {
  return {
    path,
    text,
    source: ts.createSourceFile(path, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX),
    module: moduleOf(path, register) ?? "",
    isTest: isTestFile(path),
  };
}

export function importsOf(file: SourceFile, register: Register): ImportRef[] {
  const out: ImportRef[] = [];
  for (const statement of file.source.statements) out.push(...statementImports(file, register, statement));
  for (const node of descendants(file.source)) {
    if (!ts.isCallExpression(node) || !isImportCall(node.expression)) continue;
    const argument = node.arguments[0];
    if (argument === undefined || !ts.isStringLiteral(argument)) continue;
    out.push(ref(file, register, argument.text, false, node));
  }
  return out;
}

function statementImports(file: SourceFile, register: Register, statement: ts.Statement): ImportRef[] {
  if (ts.isImportDeclaration(statement) && ts.isStringLiteral(statement.moduleSpecifier)) {
    return [ref(file, register, statement.moduleSpecifier.text, isTypeOnlyImport(statement.importClause), statement)];
  }
  if (ts.isImportEqualsDeclaration(statement) && ts.isExternalModuleReference(statement.moduleReference)) {
    const expression = statement.moduleReference.expression;
    if (!ts.isStringLiteral(expression)) return [];
    return [ref(file, register, expression.text, false, statement)];
  }
  if (
    ts.isExportDeclaration(statement) &&
    statement.moduleSpecifier !== undefined &&
    ts.isStringLiteral(statement.moduleSpecifier)
  ) {
    return [ref(file, register, statement.moduleSpecifier.text, isTypeOnlyExport(statement), statement)];
  }
  return [];
}

function ref(file: SourceFile, register: Register, specifier: string, typeOnly: boolean, node: ts.Node): ImportRef {
  const target = resolveSpecifier(file.path, specifier);
  return {
    specifier,
    typeOnly,
    line: lineOf(file.source, node),
    target,
    module: target === null ? null : moduleOf(target, register),
  };
}

function isTypeOnlyImport(clause: ts.ImportClause | undefined): boolean {
  // eslint-disable-next-line @typescript-eslint/no-deprecated -- this TypeScript exposes no phaseModifier yet
  return clause?.isTypeOnly === true;
}

function isTypeOnlyExport(statement: ts.ExportDeclaration): boolean {
  return statement.isTypeOnly;
}

function isImportCall(expression: ts.Expression): boolean {
  if (expression.kind === ts.SyntaxKind.ImportKeyword) return true;
  return ts.isIdentifier(expression) && expression.text === "import";
}

export function isExported(node: ts.Node): boolean {
  if (!ts.canHaveModifiers(node)) return false;
  return (ts.getModifiers(node) ?? []).some((modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword);
}

export function identifiersIn(node: ts.Node): ts.Identifier[] {
  return descendantsOf(node).filter(ts.isIdentifier);
}

export function descendantsOf(node: ts.Node): ts.Node[] {
  const out: ts.Node[] = [];
  const visit = (current: ts.Node): void => {
    out.push(current);
    current.forEachChild(visit);
  };
  node.forEachChild(visit);
  return out;
}

export function lineOf(source: ts.SourceFile, node: ts.Node): number {
  return source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
}

export function descendants(source: ts.SourceFile): ts.Node[] {
  const out: ts.Node[] = [];
  const visit = (node: ts.Node): void => {
    out.push(node);
    node.forEachChild(visit);
  };
  source.forEachChild(visit);
  return out;
}
