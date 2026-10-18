import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { choicesPath, coolerDocument, coolerPath, twinBridge } from "../../../test/twinFixture";
import { behaviourOf, inCommand, inReading, pickStep } from "../../../test/twinHarness";

function lastChoices(fake: ReturnType<typeof twinBridge>) {
  return JSON.parse(fake.writes.filter((write) => write.path === choicesPath).at(-1)?.content ?? "{}") as unknown;
}

function outline() {
  return within(screen.getByRole("navigation", { name: "Device outline" }));
}

describe("choices follow the readings by what they are, not by what they are called", () => {
  it("keeps the instrument, the starting value and the effects of a reading renamed in the interface", async () => {
    const { fake } = await behaviourOf();
    const group = inReading("temperature");
    fireEvent.change(group.getByLabelText("Starts at"), { target: { value: "30" } });
    fireEvent.blur(group.getByLabelText("Starts at"));
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
    const effect = within(command.getByRole("group", { name: "Effect 1 of set_duty" }));
    fireEvent.change(effect.getByLabelText("Reading that follows"), { target: { value: `m1@${coolerPath}` } });
    fireEvent.change(effect.getByLabelText("Gain"), { target: { value: "-0.5" } });
    fireEvent.change(effect.getByLabelText("Time constant"), { target: { value: "90" } });
    pickStep(/^Interface/);
    fireEvent.click(outline().getByRole("button", { name: /^temperature/ }));
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "air" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
    pickStep(/^Behaviour/);
    const renamed = await screen.findByRole("group", { name: "Reading air" });
    expect(within(renamed).getByLabelText("Starts at")).toHaveValue("30");
    expect(screen.queryByRole("group", { name: "Reading temperature" })).toBeNull();
    const again = within(inCommand("set_duty").getByRole("group", { name: "Effect 1 of set_duty" }));
    expect(again.getByLabelText("Reading that follows")).toHaveDisplayValue("air");
    expect(again.getByLabelText("Time constant")).toHaveValue("90");
    expect(again.getByText(/air moves toward its starting value plus -0.5 × duty/)).toBeInTheDocument();
    pickStep(/^Interface/);
    await vi.waitFor(() => {
      expect(lastChoices(fake)).toMatchObject({ kinds: [{ metrics: { air: { initial: 30 } } }] });
    });
  });

  it("keeps a half-filled effect through a trip to the interface and back, also when a reading is added", async () => {
    await behaviourOf();
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
    const effect = within(command.getByRole("group", { name: "Effect 1 of set_duty" }));
    fireEvent.change(effect.getByLabelText("Gain"), { target: { value: "3" } });
    fireEvent.click(command.getByRole("button", { name: "Add another effect" }));
    pickStep(/^Interface/);
    fireEvent.click(screen.getByRole("button", { name: "Add a reading" }));
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "pressure" } });
    fireEvent.change(screen.getByLabelText("Unit"), { target: { value: "Pa" } });
    fireEvent.change(screen.getByLabelText("Resolution"), { target: { value: "1" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
    pickStep(/^Behaviour/);
    expect(await screen.findByRole("group", { name: "Reading pressure" })).toBeInTheDocument();
    const back = inCommand("set_duty");
    expect(back.getByLabelText("Response")).toHaveValue("first_order");
    expect(back.getByRole("group", { name: "Effect 2 of set_duty" })).toBeInTheDocument();
    const first = within(back.getByRole("group", { name: "Effect 1 of set_duty" }));
    expect(first.getByLabelText("Gain")).toHaveValue("3");
    expect(first.getByLabelText("Reading that follows")).toHaveValue("");
    const options = within(first.getByLabelText("Reading that follows")).getAllByRole("option");
    expect(options.map((option) => option.textContent)).toEqual(["Choose…", "temperature", "pressure"]);
  });
});

describe("the effect of a command, in words", () => {
  it("says to add a second reading when no reading can follow", async () => {
    const fake = twinBridge();
    const alone = { ...coolerDocument, metrics: coolerDocument.metrics.filter((metric) => metric.name === "duty") };
    fake.addFile(coolerPath, alone, "k1");
    await behaviourOf(fake);
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
    expect(
      command.getByText("No reading can follow yet. Add a second reading in the Interface step."),
    ).toBeInTheDocument();
  });

  it("asks which reading an effect starts from only when two readings are set", async () => {
    const fake = twinBridge();
    const speed = {
      name: "speed",
      data_type: "number",
      unit: "rpm",
      eu_range: { min: 0, max: 10 },
      resolution: 1,
      writable: true,
    };
    const both = {
      ...coolerDocument,
      metrics: [...coolerDocument.metrics, speed],
      commands: [
        {
          name: "set_duty",
          risk_class: "medium",
          args_schema: { type: "object", properties: { duty: { type: "number" }, speed: { type: "number" } } },
        },
      ],
    };
    fake.addFile(coolerPath, both, "k1");
    await behaviourOf(fake);
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
    const effect = within(command.getByRole("group", { name: "Effect 1 of set_duty" }));
    expect(effect.queryByLabelText("Reading it sets")).toBeNull();
    fireEvent.change(command.getByLabelText("Reading set by speed"), { target: { value: "m4" } });
    const from = effect.getByLabelText("Reading it sets");
    expect(
      within(from)
        .getAllByRole("option")
        .map((option) => option.textContent),
    ).toEqual(["duty", "speed"]);
    fireEvent.change(from, { target: { value: "m4" } });
    expect(from).toHaveValue("m4");
  });
});
