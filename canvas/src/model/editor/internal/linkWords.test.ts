import { describe, expect, it } from "vitest";
import { deleteQuestion, linkParts, linkWords, nameCheck } from "./linkWords";
import { instanceNameRule } from "./instanceNameRule";

const outgoing = { outgoing: true, port: "heat out", other: "CT-1", otherPort: "heat in" };
const incoming = { outgoing: false, port: "heat in", other: "CH-1", otherPort: "heat out" };

describe("link words", () => {
  it("reads a link from the instance's own side", () => {
    expect(linkWords(outgoing)).toBe("heat out to CT-1 heat in");
    expect(linkWords(incoming)).toBe("CH-1 heat out to heat in");
  });

  it("orders the parts so the arrow points from the source to the target", () => {
    expect(linkParts(outgoing)).toEqual([
      { kind: "port", text: "heat out" },
      { kind: "arrow" },
      { kind: "other", id: "CT-1" },
      { kind: "port", text: "heat in" },
    ]);
    expect(linkParts(incoming)).toEqual([
      { kind: "other", id: "CH-1" },
      { kind: "port", text: "heat out" },
      { kind: "arrow" },
      { kind: "port", text: "heat in" },
    ]);
  });

  it("asks before deleting an instance and counts its links", () => {
    expect(deleteQuestion("CH-1", 1)).toBe("Delete CH-1 and its 1 link?");
    expect(deleteQuestion("CH-1", 3)).toBe("Delete CH-1 and its 3 links?");
  });

  it("checks a name against the rule, prefixing the rule when it does not fit", () => {
    expect(nameCheck("P-1")).toEqual({ fits: true, hint: instanceNameRule() });
    expect(nameCheck("bad name")).toEqual({ fits: false, hint: `Not a valid name. ${instanceNameRule()}` });
  });
});
