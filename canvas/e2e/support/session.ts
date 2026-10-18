import { spawn } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, rmSync } from "node:fs";
import { join } from "node:path";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { env } from "./env";
import type { Lab } from "./lab";
import { repository } from "./workspace";

const thermalBoard = join(repository, "examples", "devices", "thermal-board");
const thermalKinds = ["thermal_sensor", "indicator", "fan"];

const thermalPlant = "/#/plant?domain=domains%2Farduino-thermal.domain.json";

export function addThermalBench(lab: Lab): void {
  copyFileSync(
    join(thermalBoard, "thermal-board.domain.json"),
    join(lab.workspace, "domains", "arduino-thermal.domain.json"),
  );
  mkdirSync(join(lab.workspace, "kinds"), { recursive: true });
  for (const kind of thermalKinds) {
    copyFileSync(join(thermalBoard, `${kind}.kind.json`), join(lab.workspace, "kinds", `${kind}.kind.json`));
  }
}

export interface Bench {
  readonly host: string;
  readonly gateway: string;
  readonly children: ChildProcess[];
  close(): void;
}

export function openBench(binary: string): Bench {
  const dir = mkdtempSync("/tmp/ssc-");
  const host = join(dir, "h.sock");
  const children = [
    spawn(binary, ["plantwire", "host", "--socket", host, "--environment", "mixed", "--delay-ms", "100"], {
      stdio: "ignore",
    }),
  ];
  return {
    host,
    gateway: join(dir, "g.sock"),
    children,
    close: () => {
      children.forEach((child) => child.kill());
      rmSync(dir, { recursive: true, force: true });
    },
  };
}

export function startGateway(bench: Bench, device: string): void {
  const args = ["--standin-gateway", "--socket", bench.gateway, "--kinds-dir", thermalBoard, "--board", "uno-07"];
  bench.children.push(
    spawn(env("CANVAS_STANDIN"), [...args, "--device", `${device}:thermal_sensor:61`], { stdio: "ignore" }),
  );
}

export function runIdOf(words: string | null): string {
  const run = /run (c-[0-9a-f]{8})\./.exec(words ?? "")?.[1];
  if (run === undefined) throw new Error(`no run id in ${String(words)}`);
  return run;
}

export async function openSessionPanel(page: Page) {
  await page.goto(thermalPlant);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Arduino thermal bench twin" });
  await panel.getByRole("radio", { name: /^Connected session/ }).check();
  await expect(panel.getByLabel("Seed", { exact: true })).toBeFocused();
  return panel;
}
