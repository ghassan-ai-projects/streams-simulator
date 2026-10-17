import { vi } from "vitest";
import { FakeSource } from "./fakeWire";

type Handler = (url: string, init?: RequestInit) => Response | Promise<Response>;

export function bridgeTransport(handler: Handler) {
  const sources: FakeSource[] = [];
  const fetch = vi.fn((url: string, init?: RequestInit) => Promise.resolve(handler(url, init)));
  const events = vi.fn((): EventSource => {
    const source = new FakeSource();
    sources.push(source);
    return source as unknown as EventSource;
  });
  return { fetch, events, sources };
}
