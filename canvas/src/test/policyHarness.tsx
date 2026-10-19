import { fireEvent, screen } from "@testing-library/react";
import { success } from "./outcome";
import { check, sessionBridge } from "./sessionHarness";

export const a = `sha256:${"a".repeat(64)}`;
export const b = `sha256:${"b".repeat(64)}`;
export const file = "settings/default.policies.json";
export const field = () => screen.getByLabelText("Allowed policy digests");

export function checkedWith(policies: readonly string[]) {
  const fake = sessionBridge();
  fake.persistWrites();
  fake.actions.set("session.check", success({ ...check, policies }));
  return fake;
}

export async function checkThePlan() {
  fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
  await screen.findByRole("region", { name: "Allowed policies in effect" });
}
