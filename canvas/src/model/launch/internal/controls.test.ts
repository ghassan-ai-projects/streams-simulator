import { describe, expect, it } from "vitest";
import { pendingWords, refusalWords, speedChoices, speedValue } from "./controls";

describe("speed choices", () => {
  it("offers the presets and as fast as possible, and selects the live preset", () => {
    expect(speedChoices(60).map((choice) => choice.label)).toEqual([
      "1 s per s",
      "10 s per s",
      "60 s per s",
      "300 s per s",
      "3600 s per s",
      "As fast as possible",
    ]);
    expect(speedValue(60)).toBe("60");
    expect(speedValue(null)).toBe("0");
  });

  it("lists a rate that is not a preset in its place so the live rate is always shown", () => {
    const labels = speedChoices(120).map((choice) => choice.label);
    expect(labels).toEqual([
      "1 s per s",
      "10 s per s",
      "60 s per s",
      "120 s per s",
      "300 s per s",
      "3600 s per s",
      "As fast as possible",
    ]);
    expect(speedChoices(120).find((choice) => choice.value === speedValue(120))?.label).toBe("120 s per s");
  });
});

describe("control words", () => {
  it("says what is being done while a control call is out", () => {
    expect(pendingWords("pause")).toBe("Pausing…");
    expect(pendingWords("call")).toBe("Calling the effector…");
    expect(pendingWords(null)).toBe("");
  });

  it("quotes the director's refusal once, ending with a full stop", () => {
    expect(refusalWords("job_not_paused: the job is running. ")).toBe(
      "The simulator did not do that: job_not_paused: the job is running.",
    );
    expect(refusalWords(" ")).toBe("The simulator did not do that: it gave no reason.");
  });
});
