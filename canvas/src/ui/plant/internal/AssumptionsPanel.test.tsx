import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { failure } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";
import { provenanceBridge, miniDomainPath } from "../../../test/provenanceBridge";
import { planHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

const hash = planHash();

const twinExtra = {
  twin: { file: "twins/fan.twin.json", id: "mini", version: "0.1.0", status: "in_sync" },
  editable: {
    verdict: "read_only",
    reason: "twin",
    route: { code: "open_device_behavior", sentence: "Generated from the twin choices twins/fan.twin.json." },
  },
};

async function opened(extra: Record<string, unknown>, read: "composition" | "none" = "composition") {
  const bridge = provenanceBridge(extra, read);
  const app = mountApp(bridge, hash);
  await screen.findByRole("region", { name: /^Plant / });
  return { ...app, bridge };
}

function band() {
  return within(screen.getByRole("region", { name: "Provenance of this plant" }));
}

describe("the assumptions of the model", () => {
  const extra = { ...twinExtra };

  it("reads nothing until asked, then states readings, commands and what is not modelled, passing the twin choices", async () => {
    const { bridge } = await opened(extra, "none");
    await screen.findByRole("region", { name: "Provenance of this plant" });
    expect(bridge.calls.some((call) => call.name === "domain.assumptions")).toBe(false);
    fireEvent.click(band().getByRole("button", { name: "Model assumptions" }));
    const panel = within(await screen.findByRole("region", { name: "Assumptions of thermal_sensor" }));
    expect(bridge.calls.find((call) => call.name === "domain.assumptions")?.args).toEqual({
      path: miniDomainPath,
      twin: "twins/fan.twin.json",
    });
    expect(panel.getByText("streamsim/thermistor@1")).toBeInTheDocument();
    expect(panel.getAllByText("Suggested by the binary")).toHaveLength(2);
    expect(panel.getByText("Moves temperature toward the commanded value.")).toBeInTheDocument();
    const limits = within(screen.getByRole("region", { name: "Not modelled" }));
    expect(limits.getByText(/No calibration or hardware validation is attached/)).toBeInTheDocument();
  });

  it("shows the binary's refusal", async () => {
    const bridge = provenanceBridge(extra, "none");
    bridge.actions.set("domain.assumptions", failure(cannedFailure("domain_failed", 422, "cannot read")));
    mountApp(bridge, hash);
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Model assumptions" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("cannot read (domain_failed)");
  });

  it("has no accessibility violations with the assumptions open", async () => {
    const { view } = await opened(extra, "none");
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Model assumptions" }));
    await screen.findByRole("region", { name: "Assumptions of thermal_sensor" });
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("closes one drawer when the other opens, and closes it again on a second press", async () => {
    await opened(extra, "none");
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Model assumptions" }));
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect(screen.queryByRole("region", { name: "Model assumptions" })).toBeNull();
    expect(screen.getByRole("region", { name: "Lineage of this plant" })).toBeInTheDocument();
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect(screen.queryByRole("region", { name: "Lineage of this plant" })).toBeNull();
  });
});
