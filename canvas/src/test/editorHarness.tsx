import { fireEvent, render, within } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { useState } from "react";
import { vi } from "vitest";
import { EditorStore } from "../app/editorStore";
import { NavigationStore } from "../app/navigationStore";
import { success } from "../model/bridge";
import { compositionPlant } from "../model/composition";
import { EditorBar } from "../ui/editor/EditorBar";
import { EditorPanel } from "../ui/editor/EditorPanel";
import { FocusRequest } from "../ui/editor/panelProps";
import { useStore } from "../ui/useStore";
import { FakeBridge } from "./fakeBridge";
import { chillerPorts, towerPorts } from "./plantFixture";

export const path = "compositions/lab.composition.json";
const domainPath = "domains/lab.domain.json";
export const doc = {
  id: "lab",
  title: "Lab",
  areas: [{ name: "Cooling" }, { name: "Yard" }],
  equipment: [
    { use: "lib/chiller@1", names: ["CH-1"], area: "Cooling" },
    { use: "lib/tower@1", names: ["CT-1"], area: "Cooling" },
  ],
  connections: [{ from: "CH-1.heat_out", to: "CT-1.heat_in" }],
};

export class LabBridge extends FakeBridge {
  override action(name: string, args: Readonly<Record<string, unknown>>) {
    if (name !== "component.show") return super.action(name, args);
    this.calls.push({ name, args });
    const ports = args.ref === "lib/tower" ? towerPorts : chillerPorts;
    return Promise.resolve(success({ document: { ...ports, presets: { small: { title: "Small" } } } }));
  }
}

export function Harness({ editor }: { readonly editor: EditorStore }) {
  const state = useStore(editor);
  const [request] = useState(() => new FocusRequest());
  if (state.phase !== "ready") return <p>{state.phase}</p>;
  const plant = compositionPlant(state.doc, (component) => component);
  const navigation = new NavigationStore({ read: () => "#/", write: vi.fn(), subscribe: () => () => undefined });
  return (
    <ReactFlowProvider>
      <EditorBar draft={state} editor={editor} navigation={navigation} request={request} />
      <EditorPanel draft={state} plant={plant} editor={editor} request={request} />
    </ReactFlowProvider>
  );
}

export async function editing(composed = true) {
  const fake = new LabBridge();
  fake.files.set(path, success({ path, content: JSON.stringify(doc), etag: "c1" }));
  if (composed) fake.files.set(domainPath, success({ path: domainPath, content: "{}", etag: "d1" }));
  fake.actions.set("component.list", success({ components: [{ id: "lib/pump", kind: "equipment", title: "Pump" }] }));
  fake.actions.set("domain.compose", success({ channels: 4, faults: 2, effectors: 0 }));
  const editor = new EditorStore(fake);
  await editor.open(path);
  const view = render(<Harness editor={editor} />);
  return { fake, editor, view };
}

export function current(editor: EditorStore) {
  const state = editor.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state.doc;
}

export function choose(label: string, value: string, scope: HTMLElement = document.body) {
  fireEvent.change(within(scope).getByLabelText(label), { target: { value } });
}
