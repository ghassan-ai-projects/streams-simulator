import { act, fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { EditorStore } from "../../app/editorStore";
import { failed, success } from "../../model/bridge";
import { connectionsOf, equipmentOf } from "../../model/composition";
import { moveInstance } from "../../model/compositionEdits";
import { failure } from "../../test/fakeBridge";
import { choose, current, doc, editing, Harness, LabBridge, path } from "../../test/editorHarness";

describe("plant editor panel", () => {
  it("adds equipment from the library with its preset into an area", async () => {
    const { editor, view } = await editing();
    expect((await axe.run(view.container)).violations).toEqual([]);
    const form = screen.getByRole("form", { name: "Add equipment" });
    choose("Find equipment", "nothing like this", form);
    expect(within(form).getByText('Nothing in the library matches "nothing like this".')).toBeInTheDocument();
    choose("Find equipment", "pu", form);
    choose("Equipment", "lib/pump@1", form);
    expect(within(form).getByLabelText("Name")).toHaveValue("");
    expect(within(form).getByLabelText("Name")).toHaveAttribute("placeholder", "P-1");
    expect(within(form).getByText(/An instance name: a letter or digit, then letters, digits and/)).toBeInTheDocument();
    expect(await within(form).findByRole("option", { name: "Small" })).toBeInTheDocument();
    choose("Size", "small", form);
    choose("Area", "Yard", form);
    fireEvent.click(within(form).getByRole("button", { name: "Add to the plant" }));
    await vi.waitFor(() => {
      expect(screen.getByRole("heading", { name: "P-1" })).toHaveFocus();
    });
    expect(equipmentOf(current(editor)).at(-1)).toMatchObject({ use: "lib/pump@1", preset: "small", area: "Yard" });
    expect(screen.getByRole("status")).toHaveTextContent("Unsaved changes");
  });

  it("keeps a typed name when equipment is chosen or the plant changes, and shows the suggestion only as a placeholder", async () => {
    const { editor } = await editing();
    const form = screen.getByRole("form", { name: "Add equipment" });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "P-9" } });
    choose("Equipment", "lib/pump@1", form);
    expect(within(form).getByLabelText("Name")).toHaveValue("P-9");
    act(() => {
      editor.edit((current) => ({ ...current, title: "Lab two" }));
    });
    expect(within(form).getByLabelText("Name")).toHaveValue("P-9");
    expect(within(form).getByLabelText("Name")).toHaveAttribute("placeholder", "P-1");
    fireEvent.click(within(form).getByRole("button", { name: "Add to the plant" }));
    await vi.waitFor(() => {
      expect(equipmentOf(current(editor)).flatMap((entry) => entry.names)).toContain("P-9");
    });
  });

  it("refuses a name the binary's rule rejects, in the add and the rename forms, naming the rule", async () => {
    const { editor } = await editing();
    const form = screen.getByRole("form", { name: "Add equipment" });
    choose("Equipment", "lib/pump@1", form);
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Pump P 1" } });
    expect(within(form).getByText(/"Pump P 1" is not a valid name\. An instance name: a letter/)).toBeInTheDocument();
    fireEvent.click(within(form).getByRole("button", { name: "Add to the plant" }));
    expect(equipmentOf(current(editor)).flatMap((entry) => entry.names)).not.toContain("Pump P 1");
    act(() => {
      editor.select("CH-1");
    });
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "CH 9" } });
    expect(screen.getByText(/^Not a valid name\. An instance name/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Rename" }));
    expect(screen.getByLabelText("Name")).toHaveValue("CH 9");
    expect(equipmentOf(current(editor)).flatMap((entry) => entry.names)).toContain("CH-1");
  });

  it("renames, moves, links, unlinks and deletes the selected instance, and reverts", async () => {
    const { editor } = await editing();
    act(() => {
      editor.select("CH-1");
    });
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "CH-9" } });
    fireEvent.click(screen.getByRole("button", { name: "Rename" }));
    expect(screen.getByRole("heading", { name: "CH-9" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveFocus();
    choose("Area", "Yard");
    choose("Output", "power_out");
    choose("To", "CT-1");
    const inputs = within(screen.getByLabelText("Input")).getAllByRole("option");
    expect(inputs.map((option) => option.textContent)).toEqual(["Choose…", "power in"]);
    choose("Input", "power_in");
    fireEvent.click(screen.getByRole("button", { name: "Link" }));
    expect(connectionsOf(current(editor)).map((wire) => wire.from)).toEqual(["CH-9.heat_out", "CH-9.power_out"]);
    fireEvent.click(screen.getByRole("button", { name: "Remove link: heat out to CT-1 heat in" }));
    expect(screen.getByRole("heading", { name: "Links 1" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "CT-1" }));
    expect(screen.getByRole("heading", { name: "CT-1" })).toBeInTheDocument();
    act(() => {
      editor.select("CH-9");
    });
    fireEvent.click(screen.getByRole("button", { name: "Delete CH-9" }));
    expect(equipmentOf(current(editor))).toHaveLength(2);
    fireEvent.click(screen.getByRole("button", { name: "Delete CH-9 and its 1 link?" }));
    expect(equipmentOf(current(editor)).map((entry) => entry.names[0])).toEqual(["CT-1"]);
    expect(screen.getByRole("heading", { name: "Build" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(screen.getByRole("status")).toHaveTextContent("Unsaved changes");
    fireEvent.click(screen.getByRole("button", { name: "Discard changes" }));
    fireEvent.keyDown(document.body, { key: "z", ctrlKey: true });
    expect(equipmentOf(current(editor)).map((entry) => entry.names[0])).toEqual(["CT-1"]);
    fireEvent.click(screen.getByRole("button", { name: "Redo" }));
    expect(current(editor)).toEqual(doc);
    expect(screen.getByRole("status")).toHaveTextContent("No changes");
  });

  it("saves, shows the binary's issues on the instance and selects it from the list", async () => {
    const { fake, editor } = await editing();
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
    act(() => {
      editor.edit((doc) => ({ ...doc, title: "Lab two" }));
    });
    fake.actions.set(
      "domain.compose",
      failed({
        ...failure("composition_invalid", 422, "invalid"),
        issues: [{ pointer: "/connections/0/to", message: 'instance "CT-1" has no port "heat_in"' }],
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Needs fixing: 1 issue. Your changes are saved.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /^CH-1/ }));
    expect(screen.getByRole("heading", { name: "CH-1" })).toHaveFocus();
    expect(screen.getByRole("region", { name: "Issues on CH-1" })).toHaveTextContent('has no port "heat_in"');
    expect(screen.getByLabelText("Name")).toHaveAccessibleDescription(/has no port/);
    fake.actions.set("domain.compose", success({ channels: 4, faults: 2, effectors: 0 }));
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "CH-2" } });
    fireEvent.click(screen.getByRole("button", { name: "Rename" }));
    fireEvent.keyDown(document.body, { key: "s", ctrlKey: true });
    expect(
      await screen.findByText("Ready to run. Composed 4 channels, 2 faults, 0 effectors. Saved."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open plant" })).toBeInTheDocument();
  });

  it("lists repeated issues once with a count, says where a plant-wide one applies, and marks a stale list", async () => {
    const { fake, editor } = await editing();
    act(() => {
      editor.edit((doc) => ({ ...doc, title: "Lab two" }));
    });
    const same = { pointer: "/equipment/0/names/0", message: "not a name" };
    fake.actions.set(
      "domain.compose",
      failed({
        ...failure("composition_invalid", 422, "invalid"),
        issues: [same, same, same, { pointer: "/profiles", message: "needs more items" }],
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Needs fixing: 4 issues. Your changes are saved.")).toBeInTheDocument();
    const block = screen.getByRole("region", { name: "Issues from the binary" });
    expect(within(block).getAllByRole("listitem")).toHaveLength(2);
    expect(within(block).getByRole("button", { name: /CH-1.*not a name \(×3\)/ })).toBeInTheDocument();
    expect(within(block).getByText("in /profiles")).toBeInTheDocument();
    expect(within(block).queryByText(/from the last save/i)).not.toBeInTheDocument();
    act(() => {
      editor.edit((doc) => ({ ...doc, title: "Lab three" }));
    });
    expect(within(block).getByText(/These are from the last save/)).toBeInTheDocument();
  });

  it("adds an area inside another one, offers it for equipment, and refuses a repeated name", async () => {
    const { editor, view } = await editing();
    const form = screen.getByRole("form", { name: "Add area" });
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.change(within(form).getByLabelText("Area name"), { target: { value: "Yard" } });
    expect(within(form).getByText('There is already an area named "Yard".')).toBeInTheDocument();
    fireEvent.change(within(form).getByLabelText("Area name"), { target: { value: "Pump house" } });
    choose("Inside", "Yard", form);
    fireEvent.click(within(form).getByRole("button", { name: "Add area" }));
    expect(current(editor).areas).toEqual([...doc.areas, { name: "Pump house", parent: "Yard" }]);
    const equipment = screen.getByRole("form", { name: "Add equipment" });
    expect(within(equipment).getByRole("option", { name: "Yard › Pump house" })).toBeInTheDocument();
    expect(within(form).getByLabelText("Area name")).toHaveValue("");
  });

  it("says why the file was not saved", async () => {
    const { fake, editor } = await editing();
    act(() => {
      editor.edit((doc) => ({ ...doc, title: "Lab two" }));
    });
    fake.writeOutcomes.set(path, failed(failure("conflict", 409, "the file changed on disk")));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Not saved: the file changed on disk")).toBeInTheDocument();
    fake.writeOutcomes.delete(path);
    fireEvent.click(screen.getByRole("button", { name: "Reload from disk" }));
    expect(await screen.findByText("No changes")).toBeInTheDocument();
  });

  it("keeps the chosen instance when the output changes, and says why no instance can be linked", async () => {
    const { editor } = await editing();
    act(() => {
      editor.select("CH-1");
    });
    choose("Output", "heat_out");
    choose("To", "CT-1");
    choose("Input", "heat_in");
    choose("Output", "power_out");
    expect(screen.getByLabelText("To")).toHaveValue("CT-1");
    expect(screen.getByLabelText("Input")).toHaveValue("");
    choose("Output", "heat_out");
    expect(screen.getByLabelText("To")).toHaveValue("CT-1");
    expect(screen.getByLabelText("Input")).toHaveValue("");
  });

  it("names the missing target when nothing takes what an output gives", async () => {
    const fake = new LabBridge();
    const alone = { ...doc, equipment: [doc.equipment[0]], connections: [] };
    fake.files.set(path, success({ path, content: JSON.stringify(alone), etag: "d1" }));
    fake.actions.set("component.list", success({ components: [] }));
    const editor = new EditorStore(fake);
    await editor.open(path);
    render(<Harness editor={editor} />);
    act(() => {
      editor.select("CH-1");
    });
    expect(within(screen.getByLabelText("To")).getByRole("option")).toHaveTextContent("Choose an output first");
    choose("Output", "heat_out");
    expect(within(screen.getByLabelText("To")).getByRole("option")).toHaveTextContent("No instance takes heat yet");
    expect(within(screen.getByLabelText("Input")).getByRole("option")).toHaveTextContent("Choose an instance first");
  });

  it("deletes an instance without links at once, and asks first when it has links", async () => {
    const { editor } = await editing();
    act(() => {
      editor.edit((current) => ({ ...current, connections: [] }));
      editor.select("CT-1");
    });
    fireEvent.click(screen.getByRole("button", { name: "Delete CT-1" }));
    expect(equipmentOf(current(editor)).map((entry) => entry.names[0])).toEqual(["CH-1"]);
  });

  it("leaves an instance generated from an id template to the file", async () => {
    const fake = new LabBridge();
    const templated = {
      ...doc,
      equipment: [...doc.equipment, { use: "lib/tower@1", id_template: "T-{n}", area: "Yard" }],
    };
    fake.files.set(path, success({ path, content: JSON.stringify(templated), etag: "c1" }));
    fake.actions.set("component.list", success({ components: [] }));
    const editor = new EditorStore(fake);
    await editor.open(path);
    render(<Harness editor={editor} />);
    act(() => {
      editor.select("T-{n}");
    });
    expect(screen.getByText(/Generated from an id template/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete T-{n}" })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Name")).not.toBeInTheDocument();
  });

  it("says what undo and redo did, and what Revert's second press will do", async () => {
    const { editor } = await editing();
    act(() => {
      editor.edit((current) => moveInstance(current, "CT-1", "Yard"));
    });
    fireEvent.click(screen.getByRole("button", { name: "Undo" }));
    expect(screen.getByText("Undone. 0 more to undo.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Redo" }));
    expect(screen.getByText("Redone. 1 more to undo.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(screen.getByText(/^Press Discard changes again/)).toBeInTheDocument();
  });
});
