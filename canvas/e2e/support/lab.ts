import type { BrowserContextOptions } from "@playwright/test";
import { cpSync, mkdirSync, mkdtempSync } from "node:fs";
import { join } from "node:path";
import { pairedCookie, startBridge } from "./bridge";
import type { RunningBridge } from "./bridge";
import { env } from "./env";

export type Target = "app" | "empty" | "unpaired";

export interface Lab extends RunningBridge {
  readonly workspace: string;
  readonly binary: string;
}

export type StorageState = Exclude<BrowserContextOptions["storageState"], string | undefined>;

const unpaired: StorageState = { cookies: [], origins: [] };

function emptyWorkspace(folder: string): string {
  const workspace = join(folder, "empty-workspace");
  for (const part of ["domains", "compositions", "runs"]) {
    mkdirSync(join(workspace, part), { recursive: true });
  }
  return workspace;
}

function copiedWorkspace(folder: string): string {
  const workspace = join(folder, "lab-workspace");
  cpSync(env("CANVAS_TEMPLATE"), workspace, { recursive: true });
  return workspace;
}

export async function openLab(scratch: string, target: Target): Promise<Lab> {
  const folder = mkdtempSync(join(scratch, "lab-"));
  const workspace = target === "empty" ? emptyWorkspace(folder) : copiedWorkspace(folder);
  const binary = env("CANVAS_BINARY");
  return { ...(await startBridge(binary, workspace)), workspace, binary };
}

export function closeLab(lab: Lab): Promise<void> {
  const child = lab.process;
  if (child.exitCode !== null || child.signalCode !== null) return Promise.resolve();
  const exited = new Promise<void>((resolve) =>
    child.once("exit", () => {
      resolve();
    }),
  );
  child.kill();
  return exited;
}

export async function pairedState(lab: Lab, target: Target): Promise<StorageState> {
  if (target === "unpaired") return unpaired;
  const [pair = ""] = (await pairedCookie(lab)).split(";");
  const [name = "", ...value] = pair.split("=");
  const cookie = {
    name,
    value: value.join("="),
    domain: new URL(lab.url).hostname,
    path: "/",
    expires: -1,
    httpOnly: true,
    secure: false,
    sameSite: "Strict" as const,
  };
  return { cookies: [cookie], origins: [] };
}
