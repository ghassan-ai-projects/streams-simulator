import { describe, expect, it } from "vitest";
import { parseDomainFile } from "./domainFile";

describe("domain file", () => {
  it("reads the domain and its id", () => {
    expect(parseDomainFile('{"id":"mini","instances":[]}')).toEqual({
      domain: { id: "mini", instances: [] },
      id: "mini",
    });
  });

  it("gives no id to an id that is not text, instead of spelling it out", () => {
    expect(parseDomainFile('{"id":null}').id).toBe("");
    expect(parseDomainFile('{"id":{"a":1}}').id).toBe("");
    expect(parseDomainFile('{"id":5}').id).toBe("");
    expect(parseDomainFile("{}").id).toBe("");
  });

  it("treats a file that is not a JSON object as an empty domain", () => {
    for (const content of ["{not json", "[]", "null", '"mini"', ""]) {
      expect(parseDomainFile(content)).toEqual({ domain: {}, id: "" });
    }
  });
});
