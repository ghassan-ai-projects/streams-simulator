import { rmSync } from "node:fs";
import { builtStandIn } from "./support/standin";
import { buildFixture } from "./support/workspace";

export default async function globalSetup(): Promise<() => void> {
  const fixture = await buildFixture();
  process.env.CANVAS_TEMPLATE = fixture.workspace;
  process.env.CANVAS_BINARY = fixture.binary;
  process.env.CANVAS_STANDIN = builtStandIn();
  return () => {
    rmSync(fixture.root, { recursive: true, force: true });
  };
}
