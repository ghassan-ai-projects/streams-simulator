import { createDeviceDesk, reviewedOf } from "../app/devices";
import type { CaptureState } from "../app/devices";
import { captureResult, savedPump, savedPumpPath, serveCapture } from "./captureFixture";
import { FakeBridge } from "./fakeBridge";

export const text = { content: '{"type":"NBIRTH"}' };

export function opened(result: object = captureResult(), saved = false) {
  const bridge = new FakeBridge();
  serveCapture(bridge, result);
  if (saved) bridge.addFile(savedPumpPath, savedPump, "k3");
  return { bridge, session: createDeviceDesk(bridge.devices).capture() };
}

export function board(state: CaptureState) {
  const found = reviewedOf(state);
  if (found === null) throw new Error(state.phase);
  return found;
}
