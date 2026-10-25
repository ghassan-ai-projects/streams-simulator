import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { stubResizeObserver } from "../../../test/dom";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { assumptionsWire } from "../../../test/provenanceFixture";
import { choicesPath, domainPath } from "../../../test/twinFixture";
import { behaviourOf, createTwin, pickStep } from "../../../test/twinHarness";

async function checked() {
  stubResizeObserver();
  const opened = await behaviourOf();
  opened.fake.actions.set("domain.assumptions", success(assumptionsWire));
  createTwin();
  await screen.findByText(/^Twin created\./);
  pickStep(/^Check/);
  const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
  return { ...opened, check };
}

describe("the assumptions of the twin in its Check step", () => {
  it("reads nothing until asked, then states what the twin assumes, for the twin that was just created", async () => {
    const { check, fake } = await checked();
    await check.findByText("Plausibility: readings stay plausible in a nominal day");
    expect(fake.calls.some((call) => call.name === "domain.assumptions")).toBe(false);
    fireEvent.click(check.getByRole("button", { name: "Model assumptions" }));
    expect(await check.findByRole("region", { name: "Assumptions of thermal_sensor" })).toBeInTheDocument();
    expect(fake.calls.find((call) => call.name === "domain.assumptions")?.args).toEqual({
      path: domainPath,
      twin: choicesPath,
    });
    expect(check.getByRole("region", { name: "Not modelled" })).toHaveTextContent(
      "No calibration or hardware validation",
    );
    expect(
      check.getByText("Read from a plant generated from twin choices that are in sync with it."),
    ).toBeInTheDocument();
  });

  it("shows the binary's refusal in place", async () => {
    const { check, fake } = await checked();
    fake.actions.set("domain.assumptions", failure(cannedFailure("domain_failed", 422, "cannot read")));
    fireEvent.click(check.getByRole("button", { name: "Model assumptions" }));
    expect(await check.findByRole("alert")).toHaveTextContent("cannot read (domain_failed)");
  });

  it("has no accessibility violations", async () => {
    const { check, view } = await checked();
    fireEvent.click(check.getByRole("button", { name: "Model assumptions" }));
    await check.findByRole("region", { name: "Assumptions of thermal_sensor" });
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
