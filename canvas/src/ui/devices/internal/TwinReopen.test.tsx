import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { stubResizeObserver } from "../../../test/dom";
import { domainPath, refusal, twinBridge, withPlant } from "../../../test/twinFixture";
import { behaviourOf, createTwin, pickStep } from "../../../test/twinHarness";

function chip(name: RegExp) {
  return within(screen.getByRole("navigation", { name: "Device steps" })).getByRole("button", { name });
}

describe("a twin that already exists", () => {
  it("is shown as created when the device is reopened, offered as a plant, and replaced rather than created", async () => {
    stubResizeObserver();
    const { location, fake } = await behaviourOf(withPlant(twinBridge()));
    expect(await screen.findByText("Twin ready.")).toHaveAttribute("title", domainPath);
    expect(chip(/^Behaviour/)).toHaveTextContent("Created");
    expect(chip(/^Check/)).toHaveTextContent("Passed");
    expect(screen.queryByRole("button", { name: "Create twin" })).toBeNull();
    expect(fake.calls.filter((call) => call.name === "kind.twin")).toEqual([]);
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    await check.findByText("The twin matches its device");
    fireEvent.click(check.getByText("Technical details"));
    expect(check.getByText("sha256:t")).toBeInTheDocument();
    expect(check.queryByText("Parts")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Open as a plant" }));
    expect(location.hash).toBe(`#/plant?domain=${encodeURIComponent(domainPath)}`);
  });

  it("is replaced over the version on disk when the choices change", async () => {
    const { fake } = await behaviourOf(withPlant(twinBridge()));
    await screen.findByText("Twin ready.");
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Cooler B" } });
    expect(screen.getByRole("button", { name: "Replace twin" })).toHaveAttribute("aria-disabled", "false");
    expect(screen.queryByRole("button", { name: "Open as a plant" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Replace twin" }));
    expect(await screen.findByText("Twin ready.")).toBeInTheDocument();
    expect(fake.calls.filter((call) => call.name === "kind.twin").map((call) => call.args.etag)).toEqual(["d5"]);
  });

  it("says Replace instead of Create once the twin has been made in this session", async () => {
    await behaviourOf();
    expect(screen.getByRole("button", { name: "Create twin" })).toBeInTheDocument();
    createTwin();
    await screen.findByText("Twin ready.");
    expect(screen.getByRole("button", { name: "Replace twin" })).toBeInTheDocument();
    pickStep(/^Interface/);
    pickStep(/^Behaviour/);
    expect(await screen.findByRole("button", { name: "Replace twin" })).toBeInTheDocument();
    expect(chip(/^Behaviour/)).toHaveTextContent("Created");
  });

  it("shows the plant it found as not matching its device when it no longer does", async () => {
    const fake = withPlant(twinBridge());
    const drift = { pointer: "/kinds/0/implements", message: "differs", keyword: "implements_drift" };
    fake.actions.set("domain.validate", refusal(drift));
    await behaviourOf(fake);
    expect(await screen.findByText("Twin created, but it does not match its device.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Open as a plant" })).toBeNull();
    expect(chip(/^Check/)).toHaveTextContent("Failed");
  });
});
