import { writeFileSync } from "node:fs";
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

export default async function globalSetup(): Promise<() => void> {
  const fixture = buildFixture();
  const pairing = await startBridge(fixture.binary, fixture.workspace);
  const app = await startBridge(fixture.binary, fixture.workspace);
  const state = join(fixture.root, "paired.json");
  writeFileSync(state, storageState(app.url, await pairedCookie(app)));
  process.env.CANVAS_PAIRING_URL = pairing.url;
  process.env.CANVAS_PAIRING_CODE = pairing.code;
  process.env.CANVAS_APP_URL = app.url;
  process.env.CANVAS_APP_STATE = state;
  process.env.CANVAS_WORKSPACE = fixture.workspace;
  process.env.CANVAS_BINARY = fixture.binary;
  return () => {
    pairing.process.kill();
    app.process.kill();
  };
}
