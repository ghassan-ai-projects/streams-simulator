import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import type { Bench } from "./session";
import { listPolicies, openScriptedBench, openSessionPanel, planOneRealBoard, startGateway } from "./session";

export const cooling = { level_percent: 30 };

export function controlsOf(page: Page): Locator {
  return page.getByRole("region", { name: "Interrupts", exact: true });
}

export interface Started {
  readonly panel: Locator;
  readonly bench: Bench;
}

export async function startedSession(
  page: Page,
  lab: { readonly binary: string },
  policies: readonly string[] = [],
): Promise<Started> {
  const bench = openScriptedBench(lab.binary, [
    { id: "ref-a", afterS: 25, device: "fan-01", command: "set_cooling", args: cooling },
    { id: "ref-b", afterS: 30, device: "fan-01", command: "set_cooling", args: cooling },
  ]);
  try {
    const panel = await openSessionPanel(page);
    await planOneRealBoard(panel, bench, "zone-01");
    if (policies.length > 0) await listPolicies(panel, policies);
    await panel.getByRole("button", { name: "Start the session" }).click();
    await expect(panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/)).toBeVisible();
    startGateway(bench, "zone-01");
    await expect(panel.getByText(/announced every real board/)).toBeVisible({ timeout: 30_000 });
    return { panel, bench };
  } catch (error) {
    bench.close();
    throw error;
  }
}

export function runOfSession(words: string | null): string {
  const run = /run (c-[0-9a-f]{8})$/.exec(words ?? "")?.[1];
  if (run === undefined) throw new Error(`no run id in ${String(words)}`);
  return run;
}

export interface Box {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

export function boxOf(box: Box | null): Box {
  if (box === null) throw new Error("element has no box: it is not rendered or not visible");
  return box;
}
