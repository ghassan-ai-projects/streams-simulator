import type { Mock } from "vitest";
import { vi } from "vitest";

export type BridgeHandler = (url: string, init?: RequestInit) => Response | Promise<Response>;
type BridgeAnswer = (url: string, init?: RequestInit) => Promise<Response>;

export interface BridgeTransport {
  readonly fetch: Mock<BridgeAnswer>;
  readonly events: Mock<(url: string) => EventSource>;
  readonly listeners: Map<string, (event?: Event) => void>;
  readonly close: Mock<() => void>;
}

export function bridgeTransport(handler: BridgeHandler): BridgeTransport {
  const listeners = new Map<string, (event?: Event) => void>();
  const close = vi.fn(() => undefined);
  const events = vi.fn((): EventSource => {
    return {
      addEventListener: (name: string, fn: (event?: Event) => void) => listeners.set(name, fn),
      close,
      readyState: 2,
    } as unknown as EventSource;
  });
  const fetch = vi.fn((url: string, init?: RequestInit) => Promise.resolve(handler(url, init)));
  return { fetch, events, listeners, close };
}

export function respond(status: number, body: unknown): Response {
  return new Response(typeof body === "string" ? body : JSON.stringify(body), { status });
}
