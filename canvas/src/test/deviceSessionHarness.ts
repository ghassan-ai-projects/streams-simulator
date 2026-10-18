import { failure as failed, success } from "../app/ports";
import { createDeviceSession } from "../app/devices";
import type { Draft } from "../app/devices";
import { FakeBridge, cannedFailure } from "./fakeBridge";
import { kindCheck } from "./kindCheckFake";

export const path = "kinds/fan.kind.json";
export const fan = {
  kind: "fan",
  metrics: [{ name: "duty", data_type: "number", unit: "%", resolution: 1 }],
  commands: [{ name: "stop", args_schema: { type: "object", properties: {} } }],
};
export const valid = success({
  valid: true,
  kinds: [{ path, kind: "fan", kind_digest: "sha256:1", metrics: 1, commands: 1 }],
});

export function bridge(file: unknown = fan) {
  const fake = new FakeBridge();
  fake.addFile(path, file, "k1");
  fake.actions.set("kind.validate", valid);
  fake.handlers.set("kind.check", kindCheck);
  return fake;
}

export function ready(store: ReturnType<typeof createDeviceSession>): Draft {
  const state = store.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

export async function opened(fake = bridge()) {
  const store = createDeviceSession(fake.devices);
  await store.open(path);
  return { store, fake };
}

export function refusal(pointer: string, message: string, keyword = "pattern") {
  return failed({
    ...cannedFailure("command_failed", 422, "kind file fails its schema"),
    issues: [{ pointer, message, keyword }],
  });
}

export function checks(fake: FakeBridge, twin: boolean) {
  return fake.calls.filter((call) => call.name === "kind.check" && call.args.twin === twin);
}
