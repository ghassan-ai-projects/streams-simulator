import { fireEvent, render, screen, within } from "@testing-library/react";
import { createDeviceDesk } from "../app/devices";
import { createWorkspaceStore } from "../app/workspace";
import { CaptureScreen } from "../ui/workspace";
import { announced, captureResult, savedPump, savedPumpPath, serveCapture } from "./captureFixture";
import { FakeBridge } from "./fakeBridge";
import { navigation } from "./workspaceHarness";

export const heading = "Review what the board announced";

export function bridgeWithSavedPump(etag = "k3") {
  const bridge = new FakeBridge();
  bridge.addFile(savedPumpPath, savedPump, etag);
  return bridge;
}

export async function openCapture(result: object = captureResult([announced()]), bridge = new FakeBridge()) {
  serveCapture(bridge, result);
  const nav = navigation();
  const workspace = createWorkspaceStore(bridge.workspace);
  await workspace.open();
  const desk = createDeviceDesk(bridge.devices);
  const view = render(<CaptureScreen workspace={workspace} desk={desk} navigation={nav.store} />);
  return { bridge, nav, view, desk, workspace };
}

export function paste(text = '{"type":"NBIRTH"}') {
  fireEvent.change(screen.getByLabelText("Announcement text"), { target: { value: text } });
  fireEvent.click(screen.getByRole("button", { name: "Read the announcement" }));
}

export async function reviewed(result: object = captureResult([announced()]), bridge = new FakeBridge()) {
  const opened = await openCapture(result, bridge);
  paste();
  await screen.findByRole("heading", { level: 2, name: heading });
  return opened;
}

export function captureCalls(bridge: FakeBridge) {
  return bridge.calls.filter((call) => call.name === "kind.capture");
}

export function cardOf(name: string, at = 0) {
  const card = screen.getAllByRole("article", { name }).at(at);
  if (!card) throw new Error(`no card named ${name} at ${String(at)}`);
  return within(card);
}
