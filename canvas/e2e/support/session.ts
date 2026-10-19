import { spawn } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import type { Locator, Page } from "@playwright/test";
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

export interface ScriptedCommand {
  readonly id: string;
  readonly afterS: number;
  readonly device: string;
  readonly command: string;
  readonly args: Readonly<Record<string, unknown>>;
}

const worldStartNs = 1_767_225_600e9;
const scriptPolicy = `sha256:${"b".repeat(64)}`;

function openHost(binary: string, dir: string, extra: readonly string[]): Bench {
  const host = join(dir, "h.sock");
  const args = ["plantwire", "host", "--socket", host, "--environment", "mixed", ...extra];
  const children = [spawn(binary, args, { stdio: "ignore" })];
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

export function openBench(binary: string): Bench {
  return openHost(binary, mkdtempSync("/tmp/ssc-"), ["--delay-ms", "100"]);
}

export function openScriptedBench(binary: string, steps: readonly ScriptedCommand[]): Bench {
  const dir = mkdtempSync("/tmp/ssc-");
  const script = join(dir, "script.json");
  const issued = steps.map((step) => ({
    id: step.id,
    at_ns: worldStartNs + step.afterS * 1e9,
    device: step.device,
    command: step.command,
    args: step.args,
  }));
  writeFileSync(script, JSON.stringify(issued));
  return openHost(binary, dir, ["--delay-ms", "250", "--policy", scriptPolicy, "--script", script]);
}

export async function planOneRealBoard(
  panel: Locator,
  open: Bench,
  device: string,
  instance = "zone-01",
): Promise<void> {
  await panel.getByLabel("Host socket", { exact: true }).fill(open.host);
  await panel.getByLabel("Gateway socket", { exact: true }).fill(open.gateway);
  await panel.getByRole("button", { name: "Add a real board" }).click();
  await expect(panel.getByLabel("Real board 1 instance", { exact: true })).toBeFocused();
  await panel.getByLabel("Real board 1 instance", { exact: true }).selectOption(instance);
  await panel.getByLabel("Real board 1 board", { exact: true }).fill("uno-07");
  await panel.getByLabel("Real board 1 device", { exact: true }).fill(device);
}

export function startGateway(bench: Bench, device: string, kind = "thermal_sensor"): void {
  const args = ["--standin-gateway", "--socket", bench.gateway, "--kinds-dir", thermalBoard, "--board", "uno-07"];
  bench.children.push(spawn(env("CANVAS_STANDIN"), [...args, "--device", `${device}:${kind}:61`], { stdio: "ignore" }));
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
  const session = panel.getByRole("radio", { name: /^Connected session/ });
  await session.check();
  await expect(session).toBeFocused();
  await expect(panel.getByLabel("Seed", { exact: true })).toBeVisible();
  return panel;
}
