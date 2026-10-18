import { describe, expect, it } from "vitest";
import { isTwinKeyword, plainIssue } from "./issueWords";

const subject = (over: Partial<Parameters<typeof plainIssue>[1]> = {}) => ({
  name: "temperature",
  field: "name",
  member: "",
  label: "Name",
  ...over,
});

describe("the plain words for the binary's issues", () => {
  it.each([
    ["pattern", { field: "name" }, "Use letters, digits and . _ : - only"],
    ["pattern", { field: "unit" }, "This value is not in the accepted form"],
    ["required", { member: "name", label: "Name" }, "Name is required"],
    ["required", { member: "", field: "" }, "This is required"],
    ["type", {}, "This is the wrong kind of value"],
    ["enum", {}, "Choose one of the offered values"],
    ["const", {}, "Choose one of the offered values"],
    ["minimum", {}, "Below the lowest allowed value"],
    ["maximum", {}, "Above the highest allowed value"],
    ["exclusiveMinimum", {}, "Must be above zero"],
    ["exclusiveMaximum", {}, "Must be below the highest allowed value"],
    ["minLength", {}, "Cannot be empty"],
    ["maxLength", {}, "Too long"],
    ["minItems", {}, "Too few entries"],
    ["maxItems", {}, "Too many entries"],
    ["additionalProperties", { member: "colour" }, '"colour" is not a known setting'],
    ["additionalProperties", { member: "" }, "This setting is not known"],
    ["json", {}, "The file is not valid JSON"],
    ["kind_range_empty", {}, "The minimum must be below the maximum"],
    ["kind_digest_forbidden", {}, "A device definition does not state a digest"],
    ["domain_rule", {}, "A rule of the plant does not hold"],
    ["implements_missing", {}, "The device this plant implements is missing"],
    ["implements_drift", {}, "The twin no longer matches this device"],
    ["implements_checked", {}, "The twin does not match its device"],
    ["implements_failed", {}, "The twin does not match its device"],
    ["no_birth", {}, "No announcement from a board was found"],
    ["sealed_stream", {}, "This record is from the sealed part of a run and is never read here"],
    ["kind_digest_mismatch", {}, "The announcement does not match its own fingerprint, so it cannot be trusted"],
    ["kind_duplicate", {}, "The board announced the same device twice"],
    ["birth_too_large", {}, "The announcement is too large to read"],
    ["protocol_invalid", {}, "The board's protocol versions do not agree"],
    ["contract_unsupported", {}, "The board speaks a version of the protocol this simulator does not"],
    ["kind_draft", {}, "The announced device does not meet the rules for a device file"],
    ["twin_member_required", { member: "resolution" }, "Needs a resolution"],
    ["twin_member_required", { member: "safe_stop" }, "Needs a safe-stop choice"],
    ["twin_member_required", { member: "colour" }, "Needs the colour setting"],
    ["twin_member_required", { member: "" }, "Needs more detail"],
    ["twin_member_unsupported", {}, "A twin does not reproduce this setting"],
    ["twin_name_pattern", {}, "A twin needs a simpler name: lower-case letters, digits and _"],
    ["twin_type_unsupported", {}, "A twin reproduces number and on/off readings only"],
    ["twin_unit_required", {}, "temperature needs a unit"],
    ["twin_unit_required", { name: "" }, "This reading needs a unit"],
    ["twin_unit_unexpected", {}, "An on/off reading has no unit"],
    ["twin_range", {}, "The minimum must be below the maximum"],
    ["twin_risk_class", {}, "A twin accepts a low, medium or high risk class"],
    ["twin_not_object", {}, "This entry is not a definition"],
    ["twin_decode", {}, "A twin could not read this definition"],
    ["twin_conversion_needed", {}, "This instrument measures a different unit"],
    ["twin_kind_missing", {}, "No device file defines this device"],
    ["twin_name_collision", {}, "Two items share a name"],
    ["twin_unknown_choice", {}, "The choices name something the device does not have"],
    ["twin_choice_missing", {}, "A choice is still missing"],
    ["twin_instrument_mismatch", {}, "This instrument does not fit the reading"],
    ["twin_response_mismatch", {}, "This response does not fit the command"],
    ["twin_initial_range", {}, "The starting value is outside the range of the reading"],
    ["something_new", {}, "The binary does not accept this"],
    ["", {}, "The binary does not accept this"],
  ])("words %j in %j as %j", (keyword, over, expected) => {
    expect(plainIssue(keyword, subject(over))).toBe(expected);
  });

  it("tells a twin requirement from a schema issue by the keyword alone", () => {
    expect(isTwinKeyword("twin_unit_required")).toBe(true);
    expect(isTwinKeyword("twin_conversion_needed")).toBe(true);
    expect(isTwinKeyword("pattern")).toBe(false);
    expect(isTwinKeyword("implements_drift")).toBe(false);
    expect(isTwinKeyword("")).toBe(false);
    expect(isTwinKeyword("constructor")).toBe(false);
  });
});
