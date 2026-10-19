import { describe, expect, it, vi } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { createReviewSession } from "./reviewSession";

describe("copying text from a review", () => {
  it("hands the text to the clipboard it was given and reports whether it was copied", async () => {
    const timers = fakeClock();
    const copy = vi.fn((text: string) => Promise.resolve(text !== "refused"));
    const review = createReviewSession(new FakeBridge().review, timers, timers, { copy });
    expect(await review.copyText("sha256:abc")).toBe(true);
    expect(await review.copyText("refused")).toBe(false);
    expect(copy).toHaveBeenNthCalledWith(1, "sha256:abc");
  });

  it("reports that nothing was copied when no clipboard was given", async () => {
    const timers = fakeClock();
    const review = createReviewSession(new FakeBridge().review, timers, timers);
    expect(await review.copyText("sha256:abc")).toBe(false);
  });
});
