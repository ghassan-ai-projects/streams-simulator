import { readFileSync, writeFileSync } from "node:fs";
import { format, resolveConfig } from "prettier";
import { readRegister } from "./architecture/register.ts";
import type { ModuleSpec } from "./architecture/model.ts";

/**
 * Writes canvas/ARCHITECTURE.md from the register, so the human module table and the machine-readable map cannot drift.
 * `node scripts/write-architecture-doc.ts [--check]` verifies instead of writing.
 */

const registerPath = "architecture.json";
const docPath = "ARCHITECTURE.md";

export function moduleRows(): readonly { readonly name: string; readonly spec: ModuleSpec }[] {
  const register = readRegister(registerPath);
  return Object.entries(register.modules)
    .map(([name, spec]) => ({ name, spec }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

export function document(): string {
  const rows = moduleRows();
  const table = rows
    .map(
      ({ name, spec }) =>
        `| \`${name}\` | ${spec.kind} | ${spec.layer} | ${spec.lifetime} | ${spec.responsibility} | ${(spec.mayImport ?? []).map((entry) => `\`${entry}\``).join(", ") || "—"} | ${(spec.callers ?? []).map((entry) => `\`${entry}\``).join(", ") || "—"} |`,
    )
    .join("\n");
  return `# Canvas architecture (generated)

This table is generated from [architecture.json](architecture.json) by
\`node scripts/write-architecture-doc.ts\` and checked by \`npm run architecture\`, so the two cannot drift. It names the
modules that exist, what each may import and who calls it. The reasons and the migration order are in the canvas
architecture document of the migration's working folder.

| Module | Kind | Layer | Lifetime | Responsibility | May import | Callers |
| --- | --- | --- | --- | --- | --- | --- |
${table}
`;
}

export function currentDoc(path: string): string {
  try {
    return readFileSync(path, "utf8");
  } catch {
    return "";
  }
}

async function formatted(text: string): Promise<string> {
  const options = (await resolveConfig(docPath)) ?? {};
  return format(text, { ...options, parser: "markdown" });
}

async function main(): Promise<void> {
  const check = process.argv.includes("--check");
  const expected = await formatted(document());
  const actual = currentDoc(docPath);
  if (check) {
    const ok = actual === expected;
    process.stdout.write(
      `canvas architecture doc: ${ok ? "ok" : "out of date (run node scripts/write-architecture-doc.ts)"}\n`,
    );
    process.exitCode = ok ? 0 : 1;
    return;
  }
  writeFileSync(docPath, expected);
  process.stdout.write(`wrote ${docPath} with ${String(moduleRows().length)} modules\n`);
}

await main();
