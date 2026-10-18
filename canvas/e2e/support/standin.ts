import { existsSync } from "node:fs";
import { join } from "node:path";
import { repository } from "./workspace";

export function builtStandIn(): string {
  const binary = process.env.STANDIN_BIN ?? join(repository, "bin", "standin-gateway");
  if (!existsSync(binary)) {
    throw new Error(`no stand-in gateway at ${binary}: run make build, or set STANDIN_BIN`);
  }
  return binary;
}
