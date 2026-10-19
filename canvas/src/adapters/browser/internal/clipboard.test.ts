import { describe, expect, it, vi } from "vitest";
import { browserClipboard } from "./clipboard";

describe("browser clipboard", () => {
  it("writes the text and says it was copied", async () => {
    const writeText = vi.fn(() => Promise.resolve());
    expect(await browserClipboard({ writeText }).copy("sha256:abc")).toBe(true);
    expect(writeText).toHaveBeenCalledWith("sha256:abc");
  });

  it("says it was not copied when the browser refuses or has no clipboard", async () => {
    const refused = { writeText: () => Promise.reject(new DOMException("denied", "NotAllowedError")) };
    expect(await browserClipboard(refused).copy("x")).toBe(false);
    expect(await browserClipboard(undefined).copy("x")).toBe(false);
  });
});
