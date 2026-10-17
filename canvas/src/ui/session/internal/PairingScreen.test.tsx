import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { createSession } from "../../../app/session";
import { failure as failed } from "../../../test/outcome";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { PairingScreen } from "./PairingScreen";

async function setup(...refusals: string[]) {
  const bridge = new FakeBridge();
  bridge.pairOutcomes.push(...refusals.map((code) => failed(cannedFailure(code))));
  const session = createSession(bridge.session);
  await session.check();
  const view = render(<PairingScreen session={session} />);
  return { bridge, session, view };
}

async function submit(code: string) {
  fireEvent.change(screen.getByLabelText("Pairing code"), { target: { value: code } });
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Pair" }));
    await Promise.resolve();
  });
}

describe("pairing screen", () => {
  it("groups the code, counts characters and keeps Pair disabled until all 12 are in", async () => {
    const { view } = await setup();
    const field = screen.getByLabelText("Pairing code");
    fireEvent.change(field, { target: { value: "3s2qk7" } });
    expect(field).toHaveValue("3S2Q-K7");
    expect(screen.getByText("6 of 12 characters")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pair" })).toBeDisabled();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("shows a wrong code as an alert and keeps the field editable", async () => {
    await setup("wrong_code");
    await submit("AAAABBBBCCCC");
    expect(screen.getByRole("alert")).toHaveTextContent("That code did not match");
    expect(screen.getByLabelText("Pairing code")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByLabelText("Pairing code")).toBeEnabled();
  });

  it("disables the field after a lock", async () => {
    await setup("pairing_locked");
    await submit("AAAABBBBCCCC");
    expect(screen.getByRole("alert")).toHaveTextContent("Pairing is locked");
    expect(screen.getByLabelText("Pairing code")).toBeDisabled();
    expect(screen.getByRole("button", { name: "Pair" })).toBeDisabled();
  });

  it("renders nothing once paired", async () => {
    const { session, view } = await setup();
    await submit("AAAABBBBCCCC");
    expect(session.snapshot().phase).toBe("paired");
    expect(view.container).toBeEmptyDOMElement();
  });
});
