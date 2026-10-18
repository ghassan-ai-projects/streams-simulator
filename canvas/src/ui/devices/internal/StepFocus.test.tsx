import { fireEvent, render, screen, within } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it } from "vitest";
import { editingDevice } from "../../../test/deviceHarness";
import { behaviourOf, pickStep } from "../../../test/twinHarness";
import { FormTitle } from "./FormTitle";
import { StepFocusContext } from "./stepFocus";
import type { Pending } from "./stepFocus";

function home() {
  return screen.getByRole("button", { name: "Streams Simulator, workspace home" });
}

describe("where the focus goes", () => {
  it("lands on the heading of the form when a device is opened, so the next Tab reaches the form", async () => {
    await editingDevice();
    expect(screen.getByRole("heading", { level: 2, name: "Device" })).toHaveFocus();
    expect(screen.getByRole("heading", { level: 1, name: "fan" })).not.toHaveFocus();
  });

  it("lands on the heading of the step that was picked, whichever it is", async () => {
    await behaviourOf();
    expect(screen.getByRole("heading", { level: 2, name: "Twin" })).toHaveFocus();
    pickStep(/^Check/);
    expect(await screen.findByRole("heading", { level: 2, name: "Check" })).toHaveFocus();
    pickStep(/^Interface/);
    expect(await screen.findByRole("heading", { level: 2, name: "Device" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: /^duty/ }));
    expect(screen.getByRole("button", { name: /^duty/ })).not.toHaveFocus();
    expect(screen.getByRole("heading", { level: 2, name: "Reading" })).not.toHaveFocus();
  });
});

function Screens({ ready }: { readonly ready: boolean }) {
  const [pending, setPending] = useState<Pending>("fresh");
  return (
    <StepFocusContext value={{ pending, done: setPending }}>
      {ready ? <FormTitle id="final" words="Twin" /> : <FormTitle id="waiting" words="Not yet" interim />}
    </StepFocusContext>
  );
}

describe("a heading that appears after a screen that was only passing through", () => {
  it("takes the focus again when the form it stands for finally mounts, once", () => {
    const { rerender } = render(<Screens ready={false} />);
    expect(screen.getByRole("heading", { name: "Not yet" })).toHaveFocus();
    rerender(<Screens ready />);
    expect(screen.getByRole("heading", { name: "Twin" })).toHaveFocus();
    screen.getByRole("heading", { name: "Twin" }).blur();
    rerender(<Screens ready={false} />);
    expect(screen.getByRole("heading", { name: "Not yet" })).not.toHaveFocus();
  });
});

describe("asking before throwing work away", () => {
  it("asks before reverting, and keeps the work on request", async () => {
    await editingDevice();
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "fan2" } });
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    const question = within(screen.getByRole("group", { name: "Discard your changes?" }));
    expect(question.getByRole("button", { name: "Keep editing" })).toHaveFocus();
    fireEvent.click(question.getByRole("button", { name: "Keep editing" }));
    expect(screen.queryByRole("group", { name: "Discard your changes?" })).toBeNull();
    expect(screen.getByLabelText("Name")).toHaveValue("fan2");
    expect(screen.getByRole("button", { name: "Revert" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    fireEvent.click(screen.getByRole("button", { name: "Discard them" }));
    expect(screen.getByLabelText("Name")).toHaveValue("fan");
  });

  it("does nothing when there is nothing to revert", async () => {
    await editingDevice();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(screen.queryByRole("group", { name: "Discard your changes?" })).toBeNull();
  });
});

describe("the dialog that asks before leaving", () => {
  async function leaving() {
    const app = await editingDevice();
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "fan2" } });
    home().focus();
    fireEvent.click(home());
    return { app, dialog: screen.getByRole("dialog", { name: "Leave with unsaved changes?" }) };
  }

  it("is modal, keeps Tab inside it, and gives the focus back where it came from", async () => {
    const { dialog } = await leaving();
    expect(dialog).toHaveAttribute("aria-modal", "true");
    const stay = within(dialog).getByRole("button", { name: "Stay" });
    const save = within(dialog).getByRole("button", { name: "Save and leave" });
    expect(stay).toHaveFocus();
    save.focus();
    expect(fireEvent.keyDown(save, { key: "Tab" })).toBe(false);
    expect(stay).toHaveFocus();
    expect(fireEvent.keyDown(stay, { key: "Tab", shiftKey: true })).toBe(false);
    expect(save).toHaveFocus();
    const discard = within(dialog).getByRole("button", { name: "Discard and leave" });
    discard.focus();
    expect(fireEvent.keyDown(discard, { key: "Tab" })).toBe(true);
    fireEvent.click(stay);
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(home()).toHaveFocus();
  });
});
