import { test as base } from "@playwright/test";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { closeLab, openLab, pairedState } from "./lab";
import type { Lab, StorageState, Target } from "./lab";

interface TestFixtures {
  readonly target: Target;
  readonly lab: Lab;
  readonly storageState: StorageState;
  readonly baseURL: string;
}

interface WorkerFixtures {
  readonly scratch: string;
}

export { expect } from "@playwright/test";

export const test = base.extend<TestFixtures, WorkerFixtures>({
  scratch: [
    async ({}, provide) => {
      const folder = mkdtempSync(join(tmpdir(), "streamsim-canvas-worker-"));
      await provide(folder);
      rmSync(folder, { recursive: true, force: true });
    },
    { scope: "worker" },
  ],
  target: ["app", { option: true }],
  lab: async ({ scratch, target }, provide) => {
    const lab = await openLab(scratch, target);
    await provide(lab);
    await closeLab(lab);
  },
  storageState: async ({ lab, target }, provide) => {
    await provide(await pairedState(lab, target));
  },
  baseURL: async ({ lab }, provide) => {
    await provide(lab.url);
  },
});
