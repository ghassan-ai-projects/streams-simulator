import type { Transport } from "./http";

export function browserTransport(): Transport {
  return {
    fetch: (input, init) => window.fetch(input, init),
    events: (url) => new EventSource(url),
  };
}
