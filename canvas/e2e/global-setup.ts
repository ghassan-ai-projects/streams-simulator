import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { pairedCookie, startBridge } from "./support/bridge";
import { buildFixture } from "./support/workspace";

function storageState(url: string, setCookie: string): string {
  const [pair = ""] = setCookie.split(";");
  const [name = "", ...value] = pair.split("=");
  const host = new URL(url).hostname;
  const cookie = {
    name,
    value: value.join("="),
    domain: host,
    path: "/",
    expires: -1,
    httpOnly: true,
    secure: false,
    sameSite: "Strict",
  };
  return JSON.stringify({ cookies: [cookie], origins: [] });
}

function emptyPath(root: string): string {
  return join(root, "empty-workspace");
}

function emptyWorkspace(root: string): string {
  const workspace = emptyPath(root);
  for (const folder of ["domains", "compositions", "runs"]) {
    mkdirSync(join(workspace, folder), { recursive: true });
  }
  return workspace;
}

export default async function globalSetup(): Promise<() => void> {
  const fixture = buildFixture();
  const pairing = await startBridge(fixture.binary, fixture.workspace);
  const app = await startBridge(fixture.binary, fixture.workspace);
  const state = join(fixture.root, "paired.json");
  writeFileSync(state, storageState(app.url, await pairedCookie(app)));
  const empty = await startBridge(fixture.binary, emptyWorkspace(fixture.root));
  const emptyState = join(fixture.root, "paired-empty.json");
  writeFileSync(emptyState, storageState(empty.url, await pairedCookie(empty)));
  process.env.CANVAS_PAIRING_URL = pairing.url;
  process.env.CANVAS_PAIRING_CODE = pairing.code;
  process.env.CANVAS_APP_URL = app.url;
  process.env.CANVAS_APP_STATE = state;
  process.env.CANVAS_EMPTY_URL = empty.url;
  process.env.CANVAS_EMPTY_STATE = emptyState;
  process.env.CANVAS_EMPTY_WORKSPACE = emptyPath(fixture.root);
  process.env.CANVAS_WORKSPACE = fixture.workspace;
  process.env.CANVAS_BINARY = fixture.binary;
  return () => {
    pairing.process.kill();
    app.process.kill();
    empty.process.kill();
  };
}
