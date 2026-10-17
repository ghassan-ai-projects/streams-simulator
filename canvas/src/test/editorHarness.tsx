import { fireEvent, render, within } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { useState } from "react";
import type { EditorSession } from "../app/editor";
import { createNavigationStore } from "../app/shell";
import { success } from "../app/ports";
import { compositionPlant, shortKind } from "../model/editor";
import { EditorBar, EditorPanel, FocusRequest } from "../ui/editor";
import { useStore } from "../ui/primitives";
import { FakeBridge } from "./fakeBridge";
import { memoryLocation } from "./location";
import { chillerPorts, componentShown, towerPorts } from "./plantFixture";
import { createEditorSession } from "../app/editor";

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
  constructor() {
    super();
    this.handlers.set("component.show", ({ ref }) => {
      const ports = ref === "lib/tower" ? towerPorts : chillerPorts;
      return success(componentShown(ports, { presets: { small: { title: "Small" } } }));
    });
  }
}

export function Harness({ editor }: { readonly editor: EditorSession }) {
  const state = useStore(editor);
  const [request] = useState(() => new FocusRequest());
  const [navigation] = useState(() => createNavigationStore(memoryLocation()));
  if (state.phase !== "ready") return <p>{state.phase}</p>;
  const plant = compositionPlant(state.doc, shortKind);
  return (
    <ReactFlowProvider>
      <EditorBar draft={state} editor={editor} navigation={navigation} request={request} />
      <EditorPanel draft={state} plant={plant} editor={editor} request={request} />
    </ReactFlowProvider>
  );
}

export async function editing(composed = true) {
  const fake = new LabBridge();
  fake.addFile(path, doc, "c1");
  if (composed) fake.addFile(domainPath, "{}", "d1");
  fake.actions.set("component.list", success({ components: [{ id: "lib/pump", kind: "equipment", title: "Pump" }] }));
  fake.actions.set("domain.compose", success({ channels: 4, faults: 2, effectors: 0 }));
  const editor = createEditorSession(fake.editor);
  await editor.open(path);
  const view = render(<Harness editor={editor} />);
  return { fake, editor, view };
}

export function current(editor: EditorSession) {
  const state = editor.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state.doc;
}

export function choose(label: string, value: string, scope: HTMLElement = document.body) {
  fireEvent.change(within(scope).getByLabelText(label), { target: { value } });
}
