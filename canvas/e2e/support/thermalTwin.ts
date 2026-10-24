import { execFileSync } from "node:child_process";
import { copyFileSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import type { Lab } from "./lab";
import { repository } from "./workspace";

const source = join(repository, "internal", "testsupport", "testdata", "twin-thermal");
const kinds = ["thermal_sensor", "indicator", "fan"];
const twinDomain = "arduino-thermal-twin";

function inWorkspace(lab: Lab, ...parts: string[]): string {
  return join(lab.workspace, ...parts);
}

export function addThermalTwin(lab: Lab): void {
  mkdirSync(inWorkspace(lab, "kinds"), { recursive: true });
  mkdirSync(inWorkspace(lab, "twins"), { recursive: true });
  for (const kind of kinds)
    copyFileSync(join(source, `${kind}.kind.json`), inWorkspace(lab, "kinds", `${kind}.kind.json`));
  copyFileSync(join(source, "thermal-board.twin.json"), inWorkspace(lab, "twins", "fan.twin.json"));
  execFileSync(
    lab.binary,
    [
      "kind",
      "twin",
      ...kinds.map((kind) => inWorkspace(lab, "kinds", `${kind}.kind.json`)),
      "--choices",
      inWorkspace(lab, "twins", "fan.twin.json"),
      "--out",
      inWorkspace(lab, "domains", `${twinDomain}.domain.json`),
    ],
    { cwd: repository, stdio: "ignore" },
  );
}

export function recordTwinRun(lab: Lab, name: string): void {
  execFileSync(
    lab.binary,
    [
      "run",
      "-domains-dir",
      inWorkspace(lab, "domains"),
      "-kinds-dir",
      inWorkspace(lab, "kinds"),
      "-domain",
      twinDomain,
      "-seed",
      "7",
      "-duration",
      "120",
      "-out",
      inWorkspace(lab, "runs", name),
    ],
    { cwd: repository, stdio: "ignore" },
  );
}

interface Input {
  coef: number;
}

export function strengthenFan(lab: Lab, coef: number): void {
  const path = inWorkspace(lab, "domains", `${twinDomain}.domain.json`);
  const domain = JSON.parse(readFileSync(path, "utf8")) as {
    kinds: { dynamics?: { f1?: { inputs?: Input[] } }[] }[];
  };
  const inputs = domain.kinds[0]?.dynamics?.[0]?.f1?.inputs ?? [];
  const drive = inputs.at(-1);
  if (drive === undefined) throw new Error("the twin has no input to change");
  drive.coef = coef;
  writeFileSync(path, JSON.stringify(domain, null, 2));
}
