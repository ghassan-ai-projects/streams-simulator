import { describe, expect, it } from "vitest";
import { coolerDocument, instruments, kindOf } from "../../../test/twinFixture";
import { startingChoices } from "./twinDefaults";
import { chooseResponse, changeDrive, startAt } from "./twinEdits";
import { parseChoices, responseOf, serialiseChoices } from "./twinFile";

const cooler = kindOf(coolerDocument);

describe("the twin choices file", () => {
  it("writes the document the binary reads and reads it back", () => {
    const first = chooseResponse(startingChoices([cooler], instruments), "cooler", "set_duty", "first_order");
    const driven = changeDrive(
      first,
      { kind: "cooler", command: "set_duty", index: 0 },
      {
        kind: "cooler",
        metric: "temperature",
        gain: -0.5,
        timeConstantS: 120,
      },
    );
    const choices = startAt(driven, "cooler", "temperature", 30);
    const written = JSON.parse(serialiseChoices(choices)) as Record<string, unknown>;
    expect(written).toMatchObject({
      schema_version: "0.1",
      id: "cooler-twin",
      version: "0.1.0",
      title: "cooler twin",
      kinds: [
        {
          kind: "cooler",
          instance: "cooler-01",
          metrics: { temperature: { instrument: "streamsim/thermistor@1", initial: 30 }, duty: { instrument: "none" } },
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
    expect(serialiseChoices(choices).endsWith("\n")).toBe(true);
    expect(parseChoices(serialiseChoices(choices))).toEqual({ ...choices, extra: { version: "0.1.0" } });
  });

  it("leaves out what has not been decided, and keeps what the editor does not show", () => {
    const first = chooseResponse(startingChoices([cooler], instruments), "cooler", "set_duty", "first_order");
    const written = JSON.parse(
      serialiseChoices({ ...first, extra: { description: "Kept", site: { name: "bench" } } }),
    ) as {
      readonly description: string;
      readonly site: unknown;
      readonly kinds: readonly { readonly commands: { readonly set_duty: { readonly drives: readonly object[] } } }[];
    };
    expect(written.description).toBe("Kept");
    expect(written.site).toEqual({ name: "bench" });
    expect(written.kinds[0]?.commands.set_duty.drives[0]).toEqual({ from: "duty", kind: "", metric: "" });
    const none = chooseResponse(first, "cooler", "set_duty", "none");
    expect(JSON.parse(serialiseChoices(none))).toMatchObject({
      kinds: [{ commands: { set_duty: { response: "none" } } }],
    });
    const plain = serialiseChoices(none);
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
