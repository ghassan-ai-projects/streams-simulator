import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import axe from "axe-core";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { EditorStore } from "../../app/editorStore";
import { NavigationStore } from "../../app/navigationStore";
import { failed, success } from "../../model/bridge";
import { compositionPlant, connectionsOf, equipmentOf } from "../../model/composition";
import { FakeBridge, failure } from "../../test/fakeBridge";
import { chillerPorts, towerPorts } from "../../test/plantFixture";
import { useStore } from "../useStore";
import { EditorBar } from "./EditorBar";
import { EditorPanel } from "./EditorPanel";
import { FocusRequest } from "./panelProps";

const path = "compositions/lab.composition.json";
const doc = {
  id: "lab",
  title: "Lab",
  areas: [{ name: "Cooling" }, { name: "Yard" }],
  equipment: [
    { use: "lib/chiller@1", names: ["CH-1"], area: "Cooling" },
    { use: "lib/tower@1", names: ["CT-1"], area: "Cooling" },
  ],
  connections: [{ from: "CH-1.heat_out", to: "CT-1.heat_in" }],
};

class LabBridge extends FakeBridge {
  override action(name: string, args: Readonly<Record<string, unknown>>) {
    if (name !== "component.show") return super.action(name, args);
    this.calls.push({ name, args });
    const ports = args.ref === "lib/tower" ? towerPorts : chillerPorts;
    return Promise.resolve(success({ document: { ...ports, presets: { small: { title: "Small" } } } }));
  }
}

function Harness({ editor }: { readonly editor: EditorStore }) {
  const state = useStore(editor);
  const [request] = useState(() => new FocusRequest());
  if (state.phase !== "ready") return <p>{state.phase}</p>;
  const plant = compositionPlant(state.doc, (component) => component);
  const navigation = new NavigationStore({ read: () => "#/", write: vi.fn(), subscribe: () => () => undefined });
  return (
    <ReactFlowProvider>
      <EditorBar draft={state} editor={editor} navigation={navigation} />
      <EditorPanel draft={state} plant={plant} editor={editor} request={request} />
    </ReactFlowProvider>
  );
}

async function editing() {
  const fake = new LabBridge();
  fake.files.set(path, success({ path, content: JSON.stringify(doc), etag: "c1" }));
  fake.actions.set("component.list", success({ components: [{ id: "lib/pump", kind: "equipment", title: "Pump" }] }));
  fake.actions.set("domain.compose", success({ channels: 4, faults: 2, effectors: 0 }));
  const editor = new EditorStore(fake);
  await editor.open(path);
  const view = render(<Harness editor={editor} />);
  return { fake, editor, view };
}

function current(editor: EditorStore) {
  const state = editor.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state.doc;
}

function choose(label: string, value: string, scope: HTMLElement = document.body) {
  fireEvent.change(within(scope).getByLabelText(label), { target: { value } });
}

describe("plant editor panel", () => {
  it("adds equipment from the library with its preset into an area", async () => {
    const { editor, view } = await editing();
    expect((await axe.run(view.container)).violations).toEqual([]);
    const form = screen.getByRole("form", { name: "Add equipment" });
    choose("Component", "lib/pump@1", form);
    expect(await within(form).findByRole("option", { name: "Small" })).toBeInTheDocument();
    choose("Preset", "small", form);
    choose("Name", "P-1", form);
    choose("Area", "Yard", form);
    fireEvent.click(within(form).getByRole("button", { name: "Add to the plant" }));
    await vi.waitFor(() => {
      expect(screen.getByRole("heading", { name: "P-1" })).toHaveFocus();
    });
    expect(equipmentOf(current(editor)).at(-1)).toMatchObject({ use: "lib/pump@1", preset: "small", area: "Yard" });
    expect(screen.getByRole("status")).toHaveTextContent("Unsaved changes");
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
    choose("Input", "heat_in");
    fireEvent.click(screen.getByRole("button", { name: "Link" }));
    expect(connectionsOf(current(editor)).map((wire) => wire.from)).toEqual(["CH-9.heat_out", "CH-9.power_out"]);
    fireEvent.click(screen.getByRole("button", { name: "Remove link CH-9.heat_out to CT-1.heat_in" }));
    expect(screen.getByRole("heading", { name: "Links (1)" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Delete CH-9" }));
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
    expect(await screen.findByText("Saved, not composed: 1 issue")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /^CH-1/ }));
    expect(screen.getByRole("heading", { name: "CH-1" })).toHaveFocus();
    expect(screen.getByRole("region", { name: "Issues on CH-1" })).toHaveTextContent('has no port "heat_in"');
    expect(screen.getByLabelText("Name")).toHaveAccessibleDescription(/has no port/);
    fake.actions.set("domain.compose", success({ channels: 4, faults: 2, effectors: 0 }));
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "CH-2" } });
    fireEvent.click(screen.getByRole("button", { name: "Rename" }));
    fireEvent.keyDown(document.body, { key: "s", ctrlKey: true });
    expect(await screen.findByText("Saved. Composed 4 channels, 2 faults, 0 effectors.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open plant" })).toBeInTheDocument();
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
});
