import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { deferred } from "../../../test/deferred";
import { stubResizeObserver } from "../../../test/dom";
import { cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { coolerPath, domainPath, refusal, twinBridge } from "../../../test/twinFixture";
import { behaviourOf, createTwin, inCommand, inReading, pickStep } from "../../../test/twinHarness";

const pointer = "/kinds/0/commands/set_duty/drives/0/time_constant_s";
const madeWire = { id: "cooler-sensor-twin", digest: "d", channels: 1, faults: 1, effectors: 1, implements: [] };

function finishEffect(timeConstant: string) {
  const command = inCommand("set_duty");
  fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
  const effect = within(command.getByRole("group", { name: "Effect 1 of set_duty" }));
  fireEvent.change(effect.getByLabelText("Reading that follows"), { target: { value: `m1@${coolerPath}` } });
  fireEvent.change(effect.getByLabelText("Gain"), { target: { value: "-0.5" } });
  fireEvent.change(effect.getByLabelText("Time constant"), { target: { value: timeConstant } });
  return effect;
}

function stepChip(name: RegExp) {
  return within(screen.getByRole("navigation", { name: "Device steps" })).getByRole("button", { name });
}

describe("creating the twin from the screen", () => {
  it("creates it, says it is ready, and shows the check with the plant's file in the details", async () => {
    stubResizeObserver();
    const { fake, location, view } = await behaviourOf();
    createTwin();
    expect(await screen.findByText("Twin ready.")).toHaveAttribute("title", domainPath);
    expect(stepChip(/^Behaviour/)).toHaveTextContent("Created");
    expect(stepChip(/^Check/)).toHaveTextContent("Passed");
    expect(fake.calls.map((call) => call.name)).toContain("kind.twin");
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    for (const row of [
      "The device definition is valid",
      "Twin ready",
      "The twin matches its device",
      "The readings stay plausible",
    ]) {
      expect(await check.findByText(row)).toBeInTheDocument();
    }
    const details = check.getByText("Technical details").closest("details");
    expect(details).not.toHaveAttribute("open");
    expect(within(details as HTMLElement).getByText(domainPath)).toBeInTheDocument();
    expect(within(details as HTMLElement).getByText("sha256:t")).toBeInTheDocument();
    expect(within(details as HTMLElement).getByText("twins/cooler.twin.json")).toBeInTheDocument();
    expect(within(details as HTMLElement).getByText("3 readings · 2 commands")).toBeInTheDocument();
    expect(
      within(details as HTMLElement).getByText(/^temperature: No instrument: a plain reading$/),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Open as a plant" }));
    expect(location.hash).toBe(`#/plant?domain=${encodeURIComponent(domainPath)}`);
  });

  it("shows the check before anything is created", async () => {
    await behaviourOf();
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    expect(check.getByText(/Nothing to check yet/)).toBeInTheDocument();
    expect(check.getByText("No twin has been created yet")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Open as a plant" })).toBeNull();
  });

  it("shows that it is working, and keeps Create off meanwhile", async () => {
    const fake = twinBridge();
    const gate = deferred<Canned<unknown>>();
    fake.handlers.set("kind.twin", () => gate.promise);
    await behaviourOf(fake);
    createTwin();
    expect(await screen.findByText("Creating the twin…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    gate.resolve(success(madeWire));
    expect(await screen.findByText("Twin ready.")).toBeInTheDocument();
  });

  it("puts the binary's issue in plain words on the field it names, moves to it, and drops it when the value is corrected", async () => {
    const fake = twinBridge();
    fake.handlers.set("kind.twin", () => refusal({ pointer, message: "must be > 0", keyword: "exclusiveMinimum" }));
    await behaviourOf(fake);
    const effect = finishEffect("5");
    createTwin();
    expect(await screen.findByText("The binary refuses the choices: 1 issue.")).toBeInTheDocument();
    expect(effect.getByLabelText("Time constant")).toHaveAttribute("aria-invalid", "true");
    expect(effect.getByText("Must be above zero")).toBeInTheDocument();
    const list = within(screen.getByRole("alert", { name: "Issues" }));
    expect(list.getByText("cooler · set_duty · Time constant")).toBeInTheDocument();
    expect(list.getByText("must be > 0")).toBeInTheDocument();
    expect(effect.getByLabelText("Time constant")).toHaveFocus();
    expect(stepChip(/^Behaviour/)).toHaveTextContent("Needs attention");
    fireEvent.change(effect.getByLabelText("Time constant"), { target: { value: "30" } });
    expect(effect.queryByText("Must be above zero")).toBeNull();
    expect(screen.queryByRole("alert", { name: "Issues" })).toBeNull();
    expect(stepChip(/^Behaviour/)).toHaveTextContent("Ready");
    expect(stepChip(/^Behaviour/)).not.toHaveTextContent("Needs attention");
    expect(screen.getByText(/Choices are ready/)).toBeInTheDocument();
    fake.handlers.set("kind.twin", () => success(madeWire));
    createTwin();
    expect(await screen.findByText("Twin ready.")).toBeInTheDocument();
  });

  it("puts what the binary says of another device's reading on that device, with no way to the interface", async () => {
    const fake = twinBridge();
    const issue = {
      pointer: "/kinds/1/kind",
      message: 'needs "resolution"',
      keyword: "twin_member_required",
      member: "resolution",
    };
    fake.handlers.set("kind.twin", () => refusal(issue));
    await behaviourOf(fake);
    fireEvent.click(screen.getByRole("button", { name: "Add sensor to the twin" }));
    await screen.findByRole("group", { name: "Reading humidity" });
    createTwin();
    const section = screen.getByRole("heading", { level: 2, name: "sensor" }).closest("section");
    if (section === null) throw new Error("the sensor has no card");
    const card = within(section);
    expect(await card.findByText("Needs a resolution")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Set it in the Interface step" })).toBeNull();
    expect(within(screen.getByRole("alert", { name: "Issues" })).getByRole("listitem")).toHaveFocus();
  });

  it("lists an issue that names nothing in the form under the issues", async () => {
    const fake = twinBridge();
    fake.handlers.set("kind.twin", () => refusal({ pointer: "/kinds/7/kind", message: "no kind file defines it" }));
    await behaviourOf(fake);
    createTwin();
    const list = within(await screen.findByRole("alert", { name: "Issues" }));
    expect(list.getByText("Twin")).toBeInTheDocument();
    expect(list.getByText("The binary does not accept this")).toBeInTheDocument();
    expect(list.getByText("no kind file defines it")).toBeInTheDocument();
    expect(screen.getAllByText("no kind file defines it")).toHaveLength(1);
  });

  it("asks before replacing a plant with the same name", async () => {
    const fake = twinBridge();
    fake.handlers.set("kind.twin", () => failure(cannedFailure("conflict", 409, "changed")));
    await behaviourOf(fake);
    createTwin();
    expect(await screen.findByText(/A plant with this name already exists/)).toBeInTheDocument();
    fake.handlers.delete("kind.twin");
    fake.actions.set("kind.twin", success(madeWire));
    fireEvent.click(screen.getByRole("button", { name: "Replace it" }));
    expect(await screen.findByText("Twin ready.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Replace it" })).toBeNull();
  });

  it("says when the twin could not be made at all", async () => {
    const fake = twinBridge();
    fake.handlers.set("kind.twin", () => failure(cannedFailure("bridge_failed", 500, "broken")));
    await behaviourOf(fake);
    createTwin();
    expect(await screen.findByText("The twin could not be created: broken")).toBeInTheDocument();
  });

  it("says when the twin does not match its device, and what the plausibility check found", async () => {
    const fake = twinBridge();
    const drift = {
      pointer: "/kinds/0/implements",
      message: "differs from the kind file",
      keyword: "implements_drift",
    };
    fake.actions.set("domain.validate", refusal(drift));
    await behaviourOf(fake);
    createTwin();
    expect(await screen.findByText("Twin created, but it does not match its device.")).toBeInTheDocument();
    expect(screen.queryByText("Twin ready.")).toBeNull();
    expect(screen.queryByRole("button", { name: "Open as a plant" })).toBeNull();
    expect(stepChip(/^Check/)).toHaveTextContent("Failed");
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    expect(check.getByText("The twin does not match its device")).toBeInTheDocument();
    expect(check.getByText("The twin no longer matches this device")).toBeInTheDocument();
    expect(check.getByText("differs from the kind file")).toBeInTheDocument();
  });

  it("lists the findings of a plausibility check that found errors", async () => {
    const fake = twinBridge();
    const report = {
      errors: 1,
      warnings: 0,
      findings: [{ severity: "error", entity: "cooler-01", detail: "pinned at its bound" }],
    };
    fake.actions.set("domain.lint", failure({ ...cannedFailure("domain_lint", 422, "errors"), result: report }));
    await behaviourOf(fake);
    createTwin();
    expect(await screen.findByText("Twin created, but the plausibility check found 1 error.")).toBeInTheDocument();
    expect(screen.queryByText("Twin ready.")).toBeNull();
    expect(screen.queryByRole("button", { name: "Open as a plant" })).toBeNull();
    expect(stepChip(/^Check/)).toHaveTextContent("Failed");
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    expect(await check.findByText("The plausibility check found 1 error")).toBeInTheDocument();
    expect(check.getByText("cooler-01: pinned at its bound")).toBeInTheDocument();
  });

  it("calls the twin out of date as soon as a choice changes", async () => {
    await behaviourOf();
    createTwin();
    await screen.findByText("Twin ready.");
    const group = inReading("temperature");
    fireEvent.change(group.getByLabelText("Starts at"), { target: { value: "25" } });
    expect(screen.getByText(/Choices changed since the twin was created/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Open as a plant" })).toBeNull();
    expect(stepChip(/^Check/)).toHaveTextContent("Not run");
  });
});
