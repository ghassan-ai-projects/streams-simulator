import { success } from "../model/bridge";
import { FakeBridge } from "./fakeBridge";

export function launchBridge() {
  const fake = new FakeBridge();
  fake.actions.set("world.create", success({ world_id: "w-7" }));
  fake.actions.set("clock.advance", success({}));
  fake.actions.set("fault.inject", success({ fault_id: "f-0" }));
  fake.actions.set("run.end", success({}));
  fake.actions.set("world.destroy", success({}));
  return fake;
}
