import { execFileSync } from "node:child_process";

function git(args: readonly string[]): string | null {
  try {
    return execFileSync("git", [...args], { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim();
  } catch {
    return null;
  }
}

export function repositoryRoot(): string {
  return git(["rev-parse", "--show-toplevel"]) ?? process.cwd();
}

export function isShallowRepository(): string | null {
  return git(["rev-parse", "--is-shallow-repository"]);
}

export function introducingCommit(path: string): string | null {
  return git(["log", "--diff-filter=A", "--format=%H", "-1", "--", path]);
}

export function mergeBase(reference: string): string | null {
  for (const candidate of [reference, `origin/${reference}`]) {
    const base = git(["merge-base", candidate, "HEAD"]);
    if (base !== null && base !== "") return base;
  }
  return null;
}

export function show(rev: string, path: string): string | null {
  return git(["show", `${rev}:${path}`]);
}

export function parseJson(text: string): Record<string, unknown> {
  const value: unknown = JSON.parse(text);
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}
