import { describe, expect, it } from "vitest";
import {
  commandOrigins,
  commandStatuses,
  notAppliedReasons,
  notComparedReasons,
  toleranceSources,
  twinJudgeReasons,
  twinVerdicts,
} from "./twin";
import {
  commandStatusWords,
  judgeReasonWords,
  noDivergenceWords,
  notAppliedWords,
  notComparedWords,
  numberWords,
  originWords,
  toleranceSourceWords,
  verdictWords,
} from "./twinWords";

const keys = (table: Readonly<Record<string, unknown>>) => Object.keys(table).sort();
const sorted = (values: readonly string[]) => [...values].sort();

describe("the words of the twin report", () => {
  it.each<[string, Readonly<Record<string, unknown>>, readonly string[]]>([
    ["verdicts", verdictWords, twinVerdicts],
    ["no-divergence lines", noDivergenceWords, twinVerdicts],
    ["judge reasons", judgeReasonWords, twinJudgeReasons],
    ["command statuses", commandStatusWords, commandStatuses],
    ["command origins", originWords, commandOrigins],
    ["not-applied reasons", notAppliedWords, notAppliedReasons],
    ["not-compared reasons", notComparedWords, notComparedReasons],
    ["tolerance sources", toleranceSourceWords, toleranceSources],
  ])("has words for every one of the %s the binary can send, and no other", (_name, table, values) => {
    expect(keys(table)).toEqual(sorted(values));
  });

  it("tells the three verdicts apart by their labels", () => {
    expect(twinVerdicts.map((verdict) => verdictWords[verdict].label)).toEqual(["Agrees", "Diverges", "Cannot judge"]);
  });

  it("writes numbers with at most four significant digits", () => {
    expect(numberWords(2.5)).toBe("2.5");
    expect(numberWords(33.912345)).toBe("33.91");
    expect(numberWords(0)).toBe("0");
    expect(numberWords(1234567)).toBe("1235000");
  });
});
