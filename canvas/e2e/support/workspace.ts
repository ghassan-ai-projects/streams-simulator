import { spawn } from "node:child_process";
import { appendFileSync, copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";

export const repository = resolve(import.meta.dirname, "..", "..", "..");

export interface Fixture {
  readonly binary: string;
  readonly workspace: string;
  readonly root: string;
}

function run(binary: string, args: readonly string[]): Promise<void> {
  return new Promise((resolve, reject) => {
    const child = spawn(binary, [...args], { cwd: repository, stdio: ["ignore", "ignore", "pipe"] });
    let errors = "";
    child.stderr.on("data", (chunk: Buffer) => {
      errors += chunk.toString();
    });
    child.on("error", reject);
    child.on("close", (status) => {
      if (status === 0) resolve();
      else reject(new Error(`${binary} ${args.join(" ")} exited with ${String(status)}: ${errors}`));
    });
  });
}

function record(
  binary: string,
  workspace: string,
  name: string,
  extra: readonly string[],
  domain = "factory-utilities",
): Promise<void> {
  const domains = join(workspace, "domains");
  return run(binary, [
    "run",
    "-domains-dir",
    domains,
    "-domain",
    domain,
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

function markIncomplete(path: string): void {
  const artifact = JSON.parse(readFileSync(path, "utf8")) as Record<string, unknown>;
  writeFileSync(path, JSON.stringify({ ...artifact, incomplete: true }, null, 2));
}

function dropLedgerDigest(path: string): void {
  const artifact = JSON.parse(readFileSync(path, "utf8")) as Record<string, unknown>;
  delete artifact.ledger_digest;
  writeFileSync(path, JSON.stringify(artifact, null, 2));
}

function compose(binary: string, workspace: string, name: string, source: string): Promise<void> {
  const composition = join(workspace, "compositions", `${name}.composition.json`);
  copyFileSync(source, composition);
  return run(binary, ["domain", "compose", composition, "--out", join(workspace, "domains", `${name}.domain.json`)]);
}

async function composeDomains(binary: string, workspace: string): Promise<void> {
  const fixtures = join(import.meta.dirname, "..", "fixtures");
  copyFileSync(join(fixtures, "alarm-playback.domain.json"), join(workspace, "domains", "alarm-playback.domain.json"));
  copyFileSync(
    join(repository, "examples", "chiller-pilot.domain.json"),
    join(workspace, "domains", "chiller-pilot.domain.json"),
  );
  await Promise.all([
    compose(
      binary,
      workspace,
      "factory-utilities",
      join(repository, "examples", "compositions", "factory-utilities.composition.json"),
    ),
    compose(binary, workspace, "campus-3", join(fixtures, "campus-3.composition.json")),
    compose(binary, workspace, "boiler-steam-loop", join(fixtures, "boiler-steam-loop.composition.json")),
  ]);
}

async function recordRuns(binary: string, workspace: string): Promise<void> {
  const jams = ["-fault", "h02/M-1=molder__product_jam@300", "-fault", "h02/M-2=molder__bearing_wear@300"];
  const stop = 'air_dryer__set_running@D-1@900@{"dryer_id":"D-1","running":0}';
  const running = 'boiler__set_running@boiler-1@600@{"boiler_id":"boiler-1","running":1}';
  const scaling = ["-fault", "boiler-1=boiler__scaling@600", "-effector", running];
  const late = ["-fault", "D-1=air_dryer__prefilter_clogging@600", "-effector", stop];
  const plain = ["w-evaluate", "w-edited", "w-tampered", "w-incomplete-stamp", "w-missing-digest"];
  await Promise.all([
    record(binary, workspace, "w-alarm", [], "alarm-playback"),
    record(binary, workspace, "w-explore", ["-explore"]),
    record(binary, workspace, "w-hall", ["-explore", ...jams], "campus-3"),
    record(binary, workspace, "w-late", late),
    record(binary, workspace, "w-boiler", scaling, "boiler-steam-loop"),
    record(binary, workspace, "w-diverged", ["-fault", "CT-1=cooling_tower__fan_failure@600"]),
    ...plain.map((name) => record(binary, workspace, name, [])),
  ]);
}

function tamperRuns(workspace: string): void {
  const runs = join(workspace, "runs");
  appendFileSync(join(runs, "w-edited", "trace.jsonl"), "\n");
  dropCommands(join(runs, "w-diverged", "run.json"), "fault.inject");
  retitleRecordedPlant(join(runs, "w-tampered", "run.json"));
  markIncomplete(join(runs, "w-incomplete-stamp", "run.json"));
  dropLedgerDigest(join(runs, "w-missing-digest", "run.json"));
  editDescription(join(workspace, "domains", "factory-utilities.domain.json"));
}

export async function buildFixture(): Promise<Fixture> {
  const root = mkdtempSync(join(tmpdir(), "streamsim-canvas-e2e-"));
  const binary = builtBinary();
  const workspace = join(root, "lab-workspace");
  for (const folder of ["domains", "compositions", "runs"]) {
    mkdirSync(join(workspace, folder), { recursive: true });
  }
  await composeDomains(binary, workspace);
  await recordRuns(binary, workspace);
  tamperRuns(workspace);
  return { binary, workspace, root };
}
