import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { choicesPath, coolerDocument, coolerPath, twinBridge } from "../../../test/twinFixture";
import { behaviourOf, inCommand, inReading } from "../../../test/twinHarness";

function lastChoices(fake: ReturnType<typeof twinBridge>) {
  return JSON.parse(fake.writes.filter((write) => write.path === choicesPath).at(-1)?.content ?? "{}") as unknown;
}

describe("the behaviour step", () => {
  it("shows the steps with their state and the choices of the device, with no accessibility violation", async () => {
    const { view } = await behaviourOf();
    const steps = within(screen.getByRole("navigation", { name: "Device steps" }));
    expect(steps.getByRole("button", { name: /^Interface/ })).toHaveTextContent("Valid");
    expect(steps.getByRole("button", { name: /^Behaviour/ })).toHaveAttribute("aria-current", "step");
    expect(steps.getByRole("button", { name: /^Behaviour/ })).toHaveTextContent("Not started");
    expect(steps.getByRole("button", { name: /^Check/ })).toHaveTextContent("Not run");
    expect(screen.getByLabelText("Twin name")).toHaveValue("cooler-twin");
    expect(screen.getByLabelText("Title")).toHaveValue("cooler twin");
    expect(inReading("temperature").getByLabelText("Instrument")).toHaveDisplayValue(/Thermistor · degC/);
    expect(inReading("duty").getByLabelText("Instrument")).toHaveDisplayValue("No instrument: a plain reading");
    expect(inReading("mode").getByText(/A plain reading: the twin reports it as it is/)).toBeInTheDocument();
    expect(inCommand("set_duty").getByLabelText("Response")).toHaveValue("direct");
    expect(inCommand("set_duty").getByLabelText("duty sets")).toHaveValue("duty");
    expect(screen.getByText("stop: a safe stop. The twin handles it itself.")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("changes the instrument of a reading from a searchable list showing its unit and range, and saves the choice", async () => {
    const { fake } = await behaviourOf();
    const group = inReading("temperature");
    fireEvent.click(group.getByRole("button", { name: "Change the instrument of temperature" }));
    const search = group.getByLabelText("Find an instrument for temperature");
    expect(search).toHaveFocus();
    expect(group.getByText("3 instruments")).toBeInTheDocument();
    fireEvent.change(search, { target: { value: "humid" } });
    expect(group.getByText("1 instrument")).toBeInTheDocument();
    const list = group.getByRole("listbox", { name: "Instruments for temperature" });
    expect(
      within(list)
        .getAllByRole("option")
        .map((option) => option.textContent),
    ).toEqual(["No instrument: a plain reading", "Capacitive humidity sensor · %RH · 0 to 100"]);
    fireEvent.change(list, { target: { value: "streamsim/humidity@1" } });
    fireEvent.click(group.getByRole("button", { name: "Use this instrument" }));
    expect(group.getByLabelText("Instrument")).toHaveDisplayValue(/Capacitive humidity sensor/);
    expect(group.getByRole("button", { name: "Change the instrument of temperature" })).toHaveFocus();
    await vi.waitFor(() => {
      expect(lastChoices(fake)).toMatchObject({
        kinds: [{ metrics: { temperature: { instrument: "streamsim/humidity@1" } } }],
      });
    });
  });

  it("closes the list on Escape, says when nothing matches, and chooses no instrument", async () => {
    await behaviourOf();
    const group = inReading("temperature");
    fireEvent.click(group.getByRole("button", { name: "Change the instrument of temperature" }));
    const search = group.getByLabelText("Find an instrument for temperature");
    expect(fireEvent.keyDown(search, { key: "Enter" })).toBe(false);
    fireEvent.change(search, { target: { value: "zzz" } });
    expect(group.getByText('Nothing in the library matches "zzz".')).toBeInTheDocument();
    fireEvent.change(search, { target: { value: "" } });
    fireEvent.keyDown(search, { key: "Escape" });
    expect(group.queryByRole("listbox")).toBeNull();
    expect(group.getByRole("button", { name: "Change the instrument of temperature" })).toHaveFocus();
    fireEvent.click(group.getByRole("button", { name: "Change the instrument of temperature" }));
    fireEvent.change(group.getByRole("listbox"), { target: { value: "none" } });
    fireEvent.click(group.getByRole("button", { name: "Use this instrument" }));
    expect(group.getByLabelText("Instrument")).toHaveDisplayValue("No instrument: a plain reading");
    fireEvent.click(group.getByRole("button", { name: "Change the instrument of temperature" }));
    fireEvent.click(group.getByRole("button", { name: "Close the list for temperature" }));
    expect(group.queryByRole("listbox")).toBeNull();
  });

  it("does not pick an instrument that a later search hides", async () => {
    await behaviourOf();
    const group = inReading("temperature");
    fireEvent.click(group.getByRole("button", { name: "Change the instrument of temperature" }));
    fireEvent.change(group.getByRole("listbox"), { target: { value: "streamsim/tachometer@1" } });
    fireEvent.change(group.getByLabelText("Find an instrument for temperature"), { target: { value: "humid" } });
    fireEvent.click(group.getByRole("button", { name: "Use this instrument" }));
    expect(group.getByLabelText("Instrument")).toHaveDisplayValue("No instrument: a plain reading");
  });

  it("offers only plain readings when the instrument list cannot be read", async () => {
    const fake = twinBridge();
    fake.actions.delete("component.list");
    await behaviourOf(fake);
    const group = inReading("temperature");
    expect(group.getByText(/The instrument list could not be read/)).toBeInTheDocument();
    expect(group.queryByRole("button", { name: /Change the instrument/ })).toBeNull();
    expect(group.getByLabelText("Instrument")).toHaveDisplayValue("none");
  });

  it("takes where a reading starts and saves it when the field is left", async () => {
    const { stores, fake } = await behaviourOf();
    const start = inReading("temperature").getByLabelText("Starts at");
    fireEvent.change(start, { target: { value: "20" } });
    fireEvent.blur(start);
    const state = stores.devices.twin.snapshot();
    expect(state.phase === "ready" && state.doc.kinds[0]?.metrics.temperature?.initial).toBe(20);
    await vi.waitFor(() => {
      expect(lastChoices(fake)).toMatchObject({ kinds: [{ metrics: { temperature: { initial: 20 } } }] });
    });
  });

  it("names the twin, cautions about a name the binary is likely to refuse, and keeps Create off for a name that cannot be a file", async () => {
    const { fake } = await behaviourOf();
    const name = screen.getByLabelText("Twin name");
    fireEvent.change(name, { target: { value: "Cooler" } });
    expect(screen.getByText("The binary is likely to refuse this name.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
    fireEvent.change(name, { target: { value: "bad name" } });
    expect(screen.getByText("Give the twin a name: letters, digits and - only.")).toBeInTheDocument();
    expect(name).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    fireEvent.change(name, { target: { value: "" } });
    expect(screen.getByText(/3 to 64 lower-case letters/)).toBeInTheDocument();
    fireEvent.change(name, { target: { value: "cooler-b" } });
    fireEvent.blur(name);
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Cooler B" } });
    fireEvent.blur(screen.getByLabelText("Title"));
    await vi.waitFor(() => {
      expect(lastChoices(fake)).toMatchObject({ id: "cooler-b", title: "Cooler B" });
    });
  });

  it("names a device inside the twin", async () => {
    const { fake } = await behaviourOf();
    const instance = screen.getByLabelText("Name in the twin");
    expect(instance).toHaveValue("cooler-01");
    fireEvent.change(instance, { target: { value: "cooler-02" } });
    fireEvent.blur(instance);
    await vi.waitFor(() => {
      expect(lastChoices(fake)).toMatchObject({ kinds: [{ instance: "cooler-02" }] });
    });
  });
});

describe("the response of a command", () => {
  it("asks for a complete effect when a reading follows gradually, and describes it in words", async () => {
    await behaviourOf();
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
    expect(command.getByText("Fill in every part of the effect.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    const effect = within(command.getByRole("group", { name: "Effect 1 of set_duty" }));
    expect(effect.getByLabelText("Reading it sets")).toHaveValue("duty");
    fireEvent.change(effect.getByLabelText("Reading that follows"), { target: { value: "cooler/temperature" } });
    fireEvent.change(effect.getByLabelText("Gain"), { target: { value: "-0.5" } });
    fireEvent.change(effect.getByLabelText("Time constant"), { target: { value: "120" } });
    expect(
      effect.getByText(
        "temperature moves toward its starting value plus -0.5 × duty, closing about two thirds of the gap every 120 seconds.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
    fireEvent.change(effect.getByLabelText("Reading it sets"), { target: { value: "" } });
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    fireEvent.change(effect.getByLabelText("Reading it sets"), { target: { value: "duty" } });
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
  });

  it("adds and removes effects, and clears them when the command no longer changes anything", async () => {
    await behaviourOf();
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "first_order" } });
    fireEvent.click(command.getByRole("button", { name: "Add another effect" }));
    expect(command.getByRole("group", { name: "Effect 2 of set_duty" })).toBeInTheDocument();
    fireEvent.click(
      within(command.getByRole("group", { name: "Effect 2 of set_duty" })).getByRole("button", {
        name: "Remove this effect",
      }),
    );
    expect(command.queryByRole("group", { name: "Effect 2 of set_duty" })).toBeNull();
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "none" } });
    expect(command.queryByRole("group", { name: "Effect 1 of set_duty" })).toBeNull();
    expect(command.queryByLabelText("duty sets")).toBeNull();
  });

  it("asks which reading an argument sets, and lets an argument go unused", async () => {
    await behaviourOf();
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("duty sets"), { target: { value: "" } });
    expect(command.getByText("Choose the reading this command sets.")).toBeInTheDocument();
    expect(command.getByLabelText("lease_ms sets")).toHaveValue("");
    fireEvent.change(command.getByLabelText("lease_ms sets"), { target: { value: "duty" } });
    expect(command.queryByText("Choose the reading this command sets.")).toBeNull();
  });

  it("says when a command has no number argument to set a reading from", async () => {
    const fake = twinBridge();
    const plain = { name: "set_mode", args_schema: { type: "object", properties: { m: { type: "string" } } } };
    fake.addFile(coolerPath, { ...coolerDocument, commands: [plain] }, "k1");
    await behaviourOf(fake);
    const command = inCommand("set_mode");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "direct" } });
    expect(command.getByText("This command has no number argument, so it cannot set a reading.")).toBeInTheDocument();
  });
});
