import { execFileSync } from "node:child_process";
import { appendFileSync, copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";

export const repository = resolve(import.meta.dirname, "..", "..", "..");

export interface Fixture {
  readonly binary: string;
  readonly workspace: string;
  readonly root: string;
}

function run(binary: string, args: readonly string[]): void {
  execFileSync(binary, args, { cwd: repository, stdio: ["ignore", "ignore", "pipe"] });
}

function record(binary: string, workspace: string, name: string, extra: readonly string[]): void {
  const domains = join(workspace, "domains");
  run(binary, [
    "run",
    "-domains-dir",
    domains,
    "-domain",
    "factory-utilities",
    "-seed",
    "7",
    "-duration",
    "1800",
    "-out",
    join(workspace, "runs", name),
    ...extra,
  ]);
}

function builtBinary(): string {
  const binary = process.env.STREAMSIM_BIN ?? join(repository, "bin", "streams-simulator");
  if (!existsSync(binary)) {
    throw new Error(`no streamsim binary at ${binary}: run make build, or set STREAMSIM_BIN`);
  }
  return binary;
}

function dropCommands(path: string, op: string): void {
  const artifact = JSON.parse(readFileSync(path, "utf8")) as { command_log: { op: string }[] };
  const kept = artifact.command_log.filter((command) => command.op !== op);
  writeFileSync(path, JSON.stringify({ ...artifact, command_log: kept }, null, 2));
}

function retitleRecordedPlant(path: string): void {
  const artifact = JSON.parse(readFileSync(path, "utf8")) as { domain_spec: Record<string, unknown> };
  writeFileSync(
    path,
    JSON.stringify({ ...artifact, domain_spec: { ...artifact.domain_spec, title: "Tampered" } }, null, 2),
  );
}

function editDescription(path: string): void {
  const domain = JSON.parse(readFileSync(path, "utf8")) as Record<string, unknown>;
  writeFileSync(path, JSON.stringify({ ...domain, description: "Edited after the runs were recorded." }, null, 2));
}

export function buildFixture(): Fixture {
  const root = mkdtempSync(join(tmpdir(), "streamsim-canvas-e2e-"));
  const binary = builtBinary();
  const workspace = join(root, "lab-workspace");
  for (const folder of ["domains", "compositions", "runs"]) {
    mkdirSync(join(workspace, folder), { recursive: true });
  }
  const composition = join(workspace, "compositions", "factory-utilities.composition.json");
  copyFileSync(join(repository, "examples", "compositions", "factory-utilities.composition.json"), composition);
  run(binary, ["domain", "compose", composition, "--out", join(workspace, "domains", "factory-utilities.domain.json")]);
  const campus = join(workspace, "compositions", "campus-3.composition.json");
  copyFileSync(join(import.meta.dirname, "..", "fixtures", "campus-3.composition.json"), campus);
  run(binary, ["domain", "compose", campus, "--out", join(workspace, "domains", "campus-3.domain.json")]);
  record(binary, workspace, "w-explore", ["-explore"]);
  record(binary, workspace, "w-evaluate", []);
  record(binary, workspace, "w-edited", []);
  appendFileSync(join(workspace, "runs", "w-edited", "trace.jsonl"), "\n");
  record(binary, workspace, "w-diverged", ["-fault", "CT-1=cooling_tower__fan_failure@600"]);
  dropCommands(join(workspace, "runs", "w-diverged", "run.json"), "fault.inject");
  record(binary, workspace, "w-tampered", []);
  retitleRecordedPlant(join(workspace, "runs", "w-tampered", "run.json"));
  editDescription(join(workspace, "domains", "factory-utilities.domain.json"));
  return { binary, workspace, root };
}
