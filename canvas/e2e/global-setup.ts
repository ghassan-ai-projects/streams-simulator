import { rmSync } from "node:fs";
import { buildFixture } from "./support/workspace";

export default function globalSetup(): () => void {
  const fixture = buildFixture();
  process.env.CANVAS_TEMPLATE = fixture.workspace;
  process.env.CANVAS_BINARY = fixture.binary;
  return () => {
    rmSync(fixture.root, { recursive: true, force: true });
  };
}
