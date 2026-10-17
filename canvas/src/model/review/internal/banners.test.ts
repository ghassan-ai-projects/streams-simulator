import { describe, expect, it } from "vitest";
import { report } from "../../../test/runFixture";
import type { VerifyLoad } from "./banners";
import { bannerWarnings } from "./banners";

const ready: VerifyLoad = {
  phase: "ready",
  value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" },
};
const delivered = report.counts.delivered ?? 0;

describe("banner warnings", () => {
  it("says nothing for a clean run and nothing before the report", () => {
    expect(bannerWarnings(report, ready, { phase: "ready", records: delivered })).toEqual([]);
    expect(bannerWarnings(null, ready, { phase: "failed", records: 0 })).toEqual([]);
  });

  it("carries the replay failure's diagnostic as the details", () => {
    const failed: VerifyLoad = { phase: "failed", failure: { diagnostic: { message: "replay broke" } } };
    expect(bannerWarnings(report, failed, { phase: "ready", records: delivered })).toEqual([
      { text: "Not checked: this run could not be replayed for its verdict.", details: "replay broke" },
    ]);
  });

  it("warns that changed files show as they are now", () => {
    const changed = { ...report, verification: "files_differ" };
    const [warning] = bannerWarnings(changed, ready, { phase: "ready", records: delivered });
    expect(warning?.text).toMatch(/^The run's files differ from what the run wrote/);
  });

  it("compares the playback's records with the delivered count, and only once the run ended", () => {
    const short = bannerWarnings(report, ready, { phase: "ready", records: delivered - 1 });
    expect(short.map((warning) => warning.text)).toEqual([
      `Playback has ${String(delivered - 1)} observations; the report counts ${String(delivered)} delivered.`,
    ]);
    const live = { ...report, ended: false };
    expect(bannerWarnings(live, ready, { phase: "ready", records: delivered - 1 })).toEqual([]);
    expect(bannerWarnings(live, ready, { phase: "loading", records: 0 })).toEqual([]);
  });

  it("warns that a failed playback is incomplete whether or not the run ended", () => {
    const text = `Playback is incomplete: 3 of ${String(delivered)} delivered observations arrived.`;
    expect(bannerWarnings(report, ready, { phase: "failed", records: 3 })).toEqual([{ text, details: "" }]);
    expect(bannerWarnings({ ...report, ended: false }, ready, { phase: "failed", records: 3 })).toEqual([
      { text, details: "" },
    ]);
  });
});
