import type { Bridge } from "../app/ports";
import type { BridgeFailure, Outcome } from "../model/bridge";
import { failed, parseFailure, parseFile, parsePaired, parseWorkspace, success } from "../model/bridge";

export interface Transport {
  fetch(input: string, init?: RequestInit): Promise<Response>;
  events(url: string): EventSource;
}

const unreachable: BridgeFailure = {
  status: 0,
  code: "unreachable",
  message: "The bridge did not answer. Is streamsim canvas still running?",
  issues: [],
};

async function body(response: Response): Promise<unknown> {
  try {
    return (await response.json()) as unknown;
  } catch {
    return null;
  }
}

async function call<T>(
  transport: Transport,
  url: string,
  parse: (value: unknown) => T,
  init?: RequestInit,
): Promise<Outcome<T>> {
  let response: Response;
  try {
    response = await transport.fetch(url, { credentials: "same-origin", ...init });
  } catch {
    return failed(unreachable);
  }
  const value = await body(response);
  return response.ok ? success(parse(value)) : failed(parseFailure(response.status, value));
}

function post(payload: unknown): RequestInit {
  return { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) };
}

function result(value: unknown): unknown {
  return typeof value === "object" && value !== null && "result" in value ? value.result : null;
}

export function bridgeClient(transport: Transport): Bridge {
  return {
    session: () => call(transport, "/api/session", parsePaired),
    pair: (code) => call(transport, "/api/session", parsePaired, post({ code })),
    workspace: () => call(transport, "/api/workspace", parseWorkspace),
    readFile: (path) => call(transport, `/api/file?path=${encodeURIComponent(path)}`, parseFile),
    action: (name, args) => call(transport, `/api/actions/${encodeURIComponent(name)}`, result, post({ args })),
    watchWorkspace(onChange) {
      const source = transport.events("/api/events");
      source.addEventListener("workspace", onChange);
      return () => {
        source.close();
      };
    },
  };
}

export function browserTransport(): Transport {
  return {
    fetch: (input, init) => window.fetch(input, init),
    events: (url) => new EventSource(url),
  };
}
