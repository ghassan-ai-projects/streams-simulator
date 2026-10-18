import { describe, expect, it } from "vitest";
import { cooler, coolerPath, startingChoices, suggestion } from "../../../test/twinFixture";
import { changeDrive, chooseResponse, startAt } from "./twinEdits";
import { choicesOfDocument, choicesJson, parseChoices, responseOf, serialiseNamed } from "./twinFile";
import { keyChoices, serialiseChoices } from "./twinKeys";

const members = [cooler];
const at = { kind: coolerPath, command: "c1", index: 0 };

function decided() {
  const first = chooseResponse(startingChoices(), coolerPath, "c1", "first_order");
  const driven = changeDrive(first, at, { kind: coolerPath, metric: "m1", gain: -0.5, timeConstantS: 120 });
  return startAt(driven, coolerPath, "m1", 30);
}

describe("the twin choices file", () => {
  it("writes the document the binary reads, by name, and reads it back by key", () => {
    const choices = decided();
    const written = JSON.parse(serialiseChoices(choices, members)) as Record<string, unknown>;
    expect(written).toMatchObject({
      schema_version: "0.1",
      id: "cooler-sensor-twin",
      version: "0.1.0",
      title: "Twin of cooler, sensor",
      kinds: [
        {
          kind: "cooler",
          instance: "cooler-01",
          metrics: { temperature: { instrument: "none", initial: 30 }, duty: { instrument: "none" } },
          commands: {
            set_duty: {
              response: "first_order",
              sets: { duty: "duty" },
              drives: [{ from: "duty", kind: "cooler", metric: "temperature", gain: -0.5, time_constant_s: 120 }],
            },
          },
        },
      ],
    });
    expect(serialiseChoices(choices, members).endsWith("\n")).toBe(true);
    const named = parseChoices(serialiseChoices(choices, members));
    expect(named === null ? null : keyChoices(named, members)).toEqual(choices);
  });

  it("takes everything but the name, the title and the kinds from the document it was given", () => {
    const named = choicesOfDocument(suggestion);
    expect(named.extra).toEqual({ schema_version: "0.1", version: "0.1.0" });
    expect(JSON.parse(serialiseNamed(named))).toMatchObject({ schema_version: "0.1", version: "0.1.0" });
    expect(Object.keys(choicesJson(named)).slice(0, 2)).toEqual(["id", "title"]);
  });

  it("leaves out what has not been decided, and keeps what the editor does not show", () => {
    const first = chooseResponse(startingChoices(), coolerPath, "c1", "first_order");
    const written = JSON.parse(
      serialiseChoices({ ...first, extra: { description: "Kept", site: { name: "bench" } } }, members),
    ) as {
      readonly description: string;
      readonly site: unknown;
      readonly kinds: readonly { readonly commands: { readonly set_duty: { readonly drives: readonly object[] } } }[];
    };
    expect(written.description).toBe("Kept");
    expect(written.site).toEqual({ name: "bench" });
    expect(written.kinds[0]?.commands.set_duty.drives[0]).toEqual({ from: "duty", kind: "", metric: "" });
    const none = chooseResponse(first, coolerPath, "c1", "none");
    expect(JSON.parse(serialiseChoices(none, members))).toMatchObject({
      kinds: [{ commands: { set_duty: { response: "none" } } }],
    });
    const plain = serialiseChoices(none, members);
    expect(plain).not.toContain('"sets"');
    expect(plain).not.toContain('"initial"');
  });

  it("reads a file as far as it makes sense, and refuses what is not a document", () => {
    expect(parseChoices("not json")).toBeNull();
    expect(parseChoices("[1]")).toBeNull();
    expect(parseChoices("{}")).toEqual({ id: "", title: "", kinds: [], extra: {} });
    const odd = parseChoices(
      JSON.stringify({
        id: "x",
        flavour: 1,
        kinds: [
          {
            kind: "k",
            metrics: { a: {}, b: { instrument: "u/v", initial: "no" } },
            commands: { c: { response: "weird", sets: { x: "y", z: 1 }, drives: [{ gain: "high" }] } },
          },
        ],
      }),
    );
    expect(odd?.extra).toEqual({ flavour: 1 });
    expect(odd?.kinds[0]?.metrics).toEqual({
      a: { instrument: "none", initial: null },
      b: { instrument: "u/v", initial: null },
    });
    expect(odd?.kinds[0]?.commands.c).toEqual({
      response: "none",
      sets: { x: "y" },
      drives: [{ from: "", kind: "", metric: "", gain: null, timeConstantS: null }],
    });
  });

  it("reads a response by name, and anything else as none", () => {
    expect(responseOf("direct")).toBe("direct");
    expect(responseOf("first_order")).toBe("first_order");
    expect(responseOf("none")).toBe("none");
    expect(responseOf(3)).toBe("none");
  });
});
