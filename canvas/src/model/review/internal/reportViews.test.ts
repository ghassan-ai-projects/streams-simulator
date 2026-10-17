import { describe, expect, it } from "vitest";
import { fileLine } from "./reportViews";
import { report } from "../../../test/runFixture";

describe("file line", () => {
  it("says each file-check state in words a reader need not decode", () => {
    expect(fileLine(report)).toBe("Files match the record.");
    expect(fileLine({ ...report, verification: "files_differ" })).toBe("The files changed after the run.");
    expect(fileLine({ ...report, verification: "" })).toBe("The run recorded no file digests.");
  });
});
