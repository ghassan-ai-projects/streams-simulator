import { fireEvent, screen, within } from "@testing-library/react";
import { announced, captureResult, serveCapture } from "./captureFixture";
import { FakeBridge } from "./fakeBridge";
import { navigation, show } from "./workspaceHarness";

export const heading = "Review what the board announced";

export async function openCapture(result: object = captureResult([announced()]), bridge = new FakeBridge()) {
  serveCapture(bridge, result);
  const nav = navigation();
  const view = await show(bridge, nav);
  const devices = within(screen.getByRole("region", { name: "Devices" }));
  fireEvent.click(devices.getByRole("button", { name: "Capture from a board" }));
  return { bridge, nav, view };
}

export function paste(text = '{"type":"NBIRTH"}') {
  fireEvent.change(screen.getByLabelText("Announcement text"), { target: { value: text } });
  fireEvent.click(screen.getByRole("button", { name: "Read the announcement" }));
}

export async function reviewed(result: object = captureResult([announced()])) {
  const opened = await openCapture(result);
  paste();
  await screen.findByRole("heading", { level: 2, name: heading });
  return opened;
}

export function captureCalls(bridge: FakeBridge) {
  return bridge.calls.filter((call) => call.name === "kind.capture");
}

export function cardOf(name: string) {
  return within(screen.getByRole("article", { name }));
}
