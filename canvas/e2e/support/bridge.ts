import { spawn } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import { join } from "node:path";
import { repository } from "./workspace";

export interface RunningBridge {
  readonly url: string;
  readonly code: string;
  readonly process: ChildProcess;
}

const ready = /canvas: http:\/\/(\S+)\ncanvas: pairing code (\S+)\n/;

export function startBridge(binary: string, workspace: string): Promise<RunningBridge> {
  const ui = join(repository, "canvas", "dist");
  const child = spawn(binary, ["canvas", "--workspace", workspace, "--ui-dir", ui, "--addr", "127.0.0.1:0"], {
    cwd: repository,
    stdio: ["ignore", "ignore", "pipe"],
  });
  return new Promise((resolve, reject) => {
    let output = "";
    child.stderr.on("data", (chunk: Buffer) => {
      output += chunk.toString();
      const match = ready.exec(output);
      if (match?.[1] !== undefined && match[2] !== undefined) {
        resolve({ url: `http://${match[1]}`, code: match[2], process: child });
      }
    });
    child.on("exit", (status) => {
      reject(new Error(`streamsim canvas exited with ${String(status)}: ${output}`));
    });
  });
}

export async function pairedCookie(bridge: RunningBridge): Promise<string> {
  const response = await fetch(`${bridge.url}/api/session`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code: bridge.code }),
  });
  const cookie = response.headers.get("set-cookie");
  if (!response.ok || cookie === null) {
    throw new Error(`pairing failed with ${String(response.status)}`);
  }
  return cookie;
}
