import { spaced } from "../../shared/words";

interface Subject {
  readonly name: string;
  readonly field: string;
  readonly member: string;
  readonly label: string;
}

type Words = string | ((subject: Subject) => string);

interface Entry {
  readonly twin: boolean;
  readonly words: Words;
}

const needPhrases: Readonly<Record<string, string>> = {
  kind: "a name",
  name: "a name",
  data_type: "a type",
  unit: "a unit",
  eu_range: "a range",
  resolution: "a resolution",
  writable: "a setpoint choice",
  safe_stop: "a safe-stop choice",
  risk_class: "a risk class",
  args_schema: "an arguments definition",
  metrics: "at least one reading",
  commands: "a command",
};

const nameFields = new Set(["kind", "name"]);

const generic = "The binary does not accept this";

function phraseOf(member: string): string {
  return needPhrases[member] ?? (member === "" ? "more detail" : `the ${spaced(member)} setting`);
}

function needs({ member }: Subject): string {
  return `Needs ${phraseOf(member)}`;
}

export function neededPhrase(keyword: string, member: string): string | null {
  const asks = keyword === "twin_member_required" || keyword === "twin_unit_required";
  return asks ? phraseOf(member) : null;
}

function pattern({ field }: Subject): string {
  return nameFields.has(field) ? "Use letters, digits and . _ : - only" : "This value is not in the accepted form";
}

function required({ member, field, label }: Subject): string {
  return `${member === "" && field === "" ? "This" : label} is required`;
}

function unitFor({ name }: Subject): string {
  return `${name === "" ? "This reading" : name} needs a unit`;
}

function unknownSetting({ member }: Subject): string {
  return member === "" ? "This setting is not known" : `"${member}" is not a known setting`;
}

const schema = (words: Words): Entry => ({ twin: false, words });

const twin = (words: Words): Entry => ({ twin: true, words });

const entries: Readonly<Record<string, Entry>> = {
  pattern: schema(pattern),
  required: schema(required),
  type: schema("This is the wrong kind of value"),
  enum: schema("Choose one of the offered values"),
  const: schema("Choose one of the offered values"),
  minimum: schema("Below the lowest allowed value"),
  maximum: schema("Above the highest allowed value"),
  exclusiveMinimum: schema("Must be above zero"),
  exclusiveMaximum: schema("Must be below the highest allowed value"),
  minLength: schema("Cannot be empty"),
  maxLength: schema("Too long"),
  minItems: schema("Too few entries"),
  maxItems: schema("Too many entries"),
  additionalProperties: schema(unknownSetting),
  json: schema("The file is not valid JSON"),
  kind_range_empty: schema("The minimum must be below the maximum"),
  kind_digest_forbidden: schema("A device definition does not state a digest"),
  domain_rule: schema("A rule of the plant does not hold"),
  implements_missing: schema("The device this plant implements is missing"),
  implements_drift: schema("The twin no longer matches this device"),
  implements_checked: schema("The twin does not match its device"),
  implements_failed: schema("The twin does not match its device"),
  no_birth: schema("No announcement from a board was found"),
  sealed_stream: schema("This record is from the sealed part of a run and is never read here"),
  kind_digest_mismatch: schema("The announcement does not match its own fingerprint, so it cannot be trusted"),
  kind_duplicate: schema("The board announced the same device twice"),
  birth_too_large: schema("The announcement is too large to read"),
  protocol_invalid: schema("The board's protocol versions do not agree"),
  contract_unsupported: schema("The board speaks a version of the protocol this simulator does not"),
  kind_draft: schema("The announced device does not meet the rules for a device file"),
  twin_member_required: twin(needs),
  twin_member_unsupported: twin("A twin does not reproduce this setting"),
  twin_name_pattern: twin("A twin needs a simpler name: lower-case letters, digits and _"),
  twin_type_unsupported: twin("A twin reproduces number and on/off readings only"),
  twin_unit_required: twin(unitFor),
  twin_unit_unexpected: twin("An on/off reading has no unit"),
  twin_range: twin("The minimum must be below the maximum"),
  twin_risk_class: twin("A twin accepts a low, medium or high risk class"),
  twin_not_object: twin("This entry is not a definition"),
  twin_decode: twin("A twin could not read this definition"),
  twin_conversion_needed: twin("This instrument measures a different unit"),
  twin_kind_missing: twin("No device file defines this device"),
  twin_name_collision: twin("Two items share a name"),
  twin_unknown_choice: twin("The choices name something the device does not have"),
  twin_choice_missing: twin("A choice is still missing"),
  twin_instrument_mismatch: twin("This instrument does not fit the reading"),
  twin_response_mismatch: twin("This response does not fit the command"),
  twin_initial_range: twin("The starting value is outside the range of the reading"),
};

export function isTwinKeyword(keyword: string): boolean {
  return entries[keyword]?.twin === true;
}

export function plainIssue(keyword: string, subject: Subject): string {
  const words = entries[keyword]?.words ?? generic;
  return typeof words === "string" ? words : words(subject);
}
