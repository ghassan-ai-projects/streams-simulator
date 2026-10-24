import { act, fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { createEditorSession } from "../../../app/editor";
import { cannedFailure } from "../../../test/fakeBridge";
import { doc, Harness, LabBridge, path } from "../../../test/editorHarness";
import { failure, success } from "../../../test/outcome";
import { chillerPorts, componentShown, towerPorts } from "../../../test/plantFixture";

const chiller = {
  default_preset: "medium",
  presets: { large: { title: "Large (500 kW)" }, medium: { title: "Medium (250 kW)" } },
  parameters: [
    { name: "setpoint_c", title: "Chilled-water setpoint", type: "number", unit: "degC", min: 4, max: 12, default: 7 },
    { name: "mode", title: "Control mode", type: "enum", values: ["auto", "manual"], default: "auto" },
  ],
};

const plant = {
  ...doc,
  equipment: [
    { use: "lib/chiller@1", names: ["CH-1", "CH-2"], area: "Cooling" },
    { use: "lib/tower@1", names: ["CT-1"], area: "Cooling" },
    { use: "lib/chiller@1", id_template: "CH-{n}", area: "Yard" },
  ],
  connections: [],
};

async function opened(id: string) {
  const fake = new LabBridge();
  fake.handlers.set("component.show", ({ ref }) =>
    success(ref === "lib/tower" ? componentShown(towerPorts) : componentShown(chillerPorts, chiller)),
  );
  fake.addFile(path, plant, "e0");
  fake.actions.set("component.list", success({ components: [] }));
  fake.actions.set(
    "domain.compose",
    success({ channels: 3, faults: 1, effectors: 0, digest: `sha256:${"b".repeat(64)}` }),
  );
  const editor = createEditorSession(fake.editor);
  await editor.open(path);
  const view = render(<Harness editor={editor} />);
  act(() => {
    editor.select(id);
  });
  return { fake, editor, view };
}

function equipmentRows(editor: ReturnType<typeof createEditorSession>) {
  const state = editor.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state.doc.equipment as readonly Record<string, unknown>[];
}

function panel() {
  return within(screen.getByRole("region", { name: "Size and parameters" }));
}

describe("the size and parameters of an instance", () => {
  it("shows the size, and each parameter with its title, unit, bounds, default and no value of its own", async () => {
    await opened("CH-1");
    expect(panel().getByLabelText("Size")).toHaveValue("");
    expect(panel().getByRole("option", { name: "Default: Medium (250 kW)" })).toBeInTheDocument();
    const setpoint = panel().getByLabelText("Chilled-water setpoint");
    expect(setpoint).toHaveAttribute("min", "4");
    expect(setpoint).toHaveAttribute("max", "12");
    expect(setpoint).toHaveAttribute("placeholder", "7");
    expect(setpoint).toHaveValue(null);
    expect(panel().getByText("°C")).toBeInTheDocument();
    expect(panel().getByText("4 to 12 °C · default 7 °C")).toBeInTheDocument();
    expect(panel().getByLabelText("Control mode")).toHaveValue("auto");
    expect(panel().queryByRole("button", { name: /Use the default/ })).toBeNull();
  });

  it("writes a value into the composition when the field is left, giving an instance that shared an entry its own", async () => {
    const { editor } = await opened("CH-1");
    const setpoint = panel().getByLabelText("Chilled-water setpoint");
    fireEvent.change(setpoint, { target: { value: "6.5" } });
    expect(equipmentRows(editor)).toHaveLength(3);
    fireEvent.blur(setpoint);
    expect(equipmentRows(editor)[0]).toMatchObject({ names: ["CH-2"] });
    expect(equipmentRows(editor).at(-1)).toMatchObject({ names: ["CH-1"], parameters: { setpoint_c: 6.5 } });
    expect(panel().getByText(/gives CH-1 its own entry|Use the default/)).toBeInTheDocument();
  });

  it("writes a value on Enter and goes back to the default with the button", async () => {
    const { editor } = await opened("CT-1");
    expect(screen.queryByRole("region", { name: "Size and parameters" })).toBeNull();
    act(() => {
      editor.select("CH-2");
    });
    const setpoint = panel().getByLabelText("Chilled-water setpoint");
    fireEvent.change(setpoint, { target: { value: "9" } });
    fireEvent.keyDown(setpoint, { key: "Enter" });
    expect(equipmentRows(editor).at(-1)).toMatchObject({ parameters: { setpoint_c: 9 } });
    fireEvent.click(panel().getByRole("button", { name: "Use the default of Chilled-water setpoint" }));
    expect("parameters" in (equipmentRows(editor).at(-1) ?? {})).toBe(false);
  });

  it("clears a field that is emptied and ignores text that is not a number", async () => {
    const { editor } = await opened("CH-1");
    const setpoint = panel().getByLabelText("Chilled-water setpoint");
    fireEvent.change(setpoint, { target: { value: "5" } });
    fireEvent.blur(setpoint);
    fireEvent.change(setpoint, { target: { value: "" } });
    fireEvent.blur(setpoint);
    expect(JSON.stringify(equipmentRows(editor))).not.toContain("setpoint_c");
  });

  it("chooses a size and an enumerated value", async () => {
    const { editor } = await opened("CH-1");
    fireEvent.change(panel().getByLabelText("Size"), { target: { value: "large" } });
    expect(equipmentRows(editor).at(-1)).toMatchObject({ names: ["CH-1"], preset: "large" });
    fireEvent.change(panel().getByLabelText("Control mode"), { target: { value: "manual" } });
    expect(equipmentRows(editor).at(-1)).toMatchObject({ preset: "large", parameters: { mode: "manual" } });
  });

  it("is marked unsaved like every other edit", async () => {
    await opened("CH-1");
    fireEvent.change(panel().getByLabelText("Size"), { target: { value: "large" } });
    expect(screen.getByRole("status")).toHaveTextContent("Unsaved changes");
  });

  it("says an instance made from an id template cannot be edited here", async () => {
    await opened("CH-{n}");
    expect(screen.getByText("Generated from an id template; change it in the composition file.")).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Size and parameters" })).toBeNull();
  });

  it("has no accessibility violations", async () => {
    const { view } = await opened("CH-1");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("shows the binary's refusal of a value at its field after a save, until the next change", async () => {
    const { fake, editor } = await opened("CH-1");
    fireEvent.change(panel().getByLabelText("Chilled-water setpoint"), { target: { value: "40" } });
    fireEvent.blur(panel().getByLabelText("Chilled-water setpoint"));
    const index = equipmentRows(editor).length - 1;
    fake.actions.set(
      "domain.compose",
      failure({
        ...cannedFailure("composition_invalid", 422, "refused"),
        issues: [
          {
            pointer: `/equipment/${String(index)}/parameters/setpoint_c`,
            message: '40 is outside the bounds of parameter "setpoint_c"',
          },
        ],
      }),
    );
    await act(async () => {
      await editor.save();
    });
    expect(panel().getByRole("alert")).toHaveTextContent('40 is outside the bounds of parameter "setpoint_c"');
    fireEvent.change(panel().getByLabelText("Control mode"), { target: { value: "manual" } });
    expect(panel().queryByRole("alert")).toBeNull();
  });

  it("shows the binary's refusal of a size at the size field", async () => {
    const { fake, editor } = await opened("CH-1");
    fireEvent.change(panel().getByLabelText("Size"), { target: { value: "large" } });
    const index = equipmentRows(editor).length - 1;
    fake.actions.set(
      "domain.compose",
      failure({
        ...cannedFailure("composition_invalid", 422, "refused"),
        issues: [{ pointer: `/equipment/${String(index)}/preset`, message: 'chiller has no preset "large"' }],
      }),
    );
    await act(async () => {
      await editor.save();
    });
    expect(panel().getByRole("alert")).toHaveTextContent('chiller has no preset "large"');
  });
});
