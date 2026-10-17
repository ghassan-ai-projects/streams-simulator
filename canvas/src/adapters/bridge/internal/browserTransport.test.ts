import { afterEach, describe, expect, it, vi } from "vitest";
import { browserTransport } from "..";

describe("browser transport", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("fetches through the window with the same input and init", async () => {
    const answer = new Response("{}", { status: 200 });
    const fetch = vi.spyOn(window, "fetch").mockResolvedValue(answer);
    const init = { method: "POST" };
    expect(await browserTransport().fetch("/api/session", init)).toBe(answer);
    expect(fetch).toHaveBeenCalledWith("/api/session", init);
  });

  it("opens an event stream on the given url", () => {
    const opened: string[] = [];
    class FakeSource {
      constructor(readonly url: string) {
        opened.push(url);
      }
    }
    vi.stubGlobal("EventSource", FakeSource);
    const source = browserTransport().events("/api/events");
    expect(source).toBeInstanceOf(FakeSource);
    expect(opened).toEqual(["/api/events"]);
  });
});
