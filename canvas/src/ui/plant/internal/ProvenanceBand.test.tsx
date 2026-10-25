import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { failure, success } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";
import { provenanceBridge, miniDomainPath } from "../../../test/provenanceBridge";
import { miniEditable } from "../../../test/provenanceFixture";
import { planHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

const hash = planHash();

const frame = { schema_version: "0.1", version: "1", stresses: "none", axes: {}, profiles: [], ground_truth: {} };

const hash0 = (value: string) => `sha256:${value.repeat(64)}`;

const twinExtra = {
  twin: { file: "twins/fan.twin.json", id: "mini", version: "0.1.0", status: "in_sync" },
  editable: {
    verdict: "read_only",
    reason: "twin",
    route: { code: "open_device_behavior", sentence: "Generated from the twin choices twins/fan.twin.json." },
  },
};

const exampleExtra = {
  editable: {
    verdict: "read_only",
    reason: "example",
    route: { code: "create_from_example", sentence: "A shipped example: create a plant from it under a new id." },
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

describe("the provenance of a plant produced by its composition", () => {
  const extra = {
    composition: { file: "compositions/mini.composition.json", id: "mini", version: "0.1.0", status: "in_sync" },
    runs: [
      {
        id: "w-1",
        mode: "evaluate",
        recorded: true,
        domain_version: "0.1.0",
        domain_digest: hash0("a"),
        same_as_current: true,
      },
      {
        id: "w-2",
        mode: "explore",
        recorded: true,
        domain_version: "0.1.0",
        domain_digest: hash0("b"),
        same_as_current: false,
      },
    ],
    editable: miniEditable,
  };

  it("says nothing about read-only and keeps the lineage collapsed until it is asked for", async () => {
    await opened(extra);
    expect(await screen.findByRole("region", { name: "Provenance of this plant" })).toBeInTheDocument();
    expect(band().queryByRole("region", { name: "Why this plant cannot be edited here" })).toBeNull();
    expect(band().getByRole("button", { name: "Lineage" })).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("region", { name: "Lineage of this plant" })).toBeNull();
  });

  it("lists where the plant comes from and the runs that used it, naming those that ran on another version", async () => {
    const { bridge } = await opened(extra);
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    const lineage = within(screen.getByRole("region", { name: "Lineage of this plant" }));
    expect(lineage.getByText("mini 0.1.0 ·", { exact: false })).toBeInTheDocument();
    expect(lineage.getByTitle(hash0("a"))).toHaveTextContent("sha256:aaaaaaaa…aaaaaaaa");
    const chain = within(lineage.getByRole("list", { name: "Where this plant comes from" }));
    expect(chain.getByText("compositions/mini.composition.json · in sync with this plant")).toBeInTheDocument();
    expect(lineage.getByText("same as current")).toBeInTheDocument();
    expect(lineage.getByText("ran on an earlier version of this plant")).toBeInTheDocument();
    fireEvent.click(lineage.getByRole("button", { name: "Open the review of w-1" }));
    expect(bridge.calls.filter((call) => call.name === "domain.lineage")).toHaveLength(1);
  });

  it("opens the model comparison of an earlier run from the lineage", async () => {
    const app = await opened(extra);
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect(screen.queryByRole("button", { name: "What changed since w-1" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "What changed since w-2" }));
    expect(app.location.hash).toBe("#/review?world=w-2&model=1");
  });

  it("lists the files that could not be read with the binary's message", async () => {
    await opened(extra);
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    const files = within(screen.getByRole("region", { name: "Files that could not be read" }));
    expect(files.getByText("kinds/broken.kind.json")).toBeInTheDocument();
    expect(files.getByText(/not JSON/)).toBeInTheDocument();
  });

  it("says when no run has used the plant", async () => {
    await opened({ ...extra, runs: [] });
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect(screen.getByText("No run has used this plant yet.")).toBeInTheDocument();
  });

  it("has no accessibility violations with the lineage open", async () => {
    const { view } = await opened(extra);
    await screen.findByRole("region", { name: "Provenance of this plant" });
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("keeps the strip, so the screen does not move, with the lineage off and the assumptions on when the bridge cannot give a lineage", async () => {
    const bridge = provenanceBridge(extra);
    bridge.actions.set("domain.lineage", failure(cannedFailure("unknown_action", 404)));
    mountApp(bridge, hash);
    await screen.findByRole("region", { name: /^Plant / });
    await vi.waitFor(() => {
      expect(bridge.calls.some((call) => call.name === "domain.lineage")).toBe(true);
    });
    expect(band().getByRole("button", { name: "Lineage" })).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect(screen.queryByRole("region", { name: "Lineage of this plant" })).toBeNull();
    expect(band().getByRole("button", { name: "Model assumptions" })).toHaveAttribute("aria-disabled", "false");
  });
});

describe("a plant generated from a twin", () => {
  it("says on the strip that it is read-only and opens the Behaviour step of the device that carries the twin", async () => {
    const app = await opened(twinExtra, "none");
    const notice = await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    expect(within(notice).getByText("Read-only: this plant is generated from a twin.")).toBeInTheDocument();
    fireEvent.click(within(notice).getByRole("button", { name: "Open the twin's Behaviour step" }));
    expect(app.location.hash).toBe("#/device?device=kinds%2Ffan.kind.json&step=behaviour");
  });

  it("gives the binary's sentence and how the links read in the lineage drawer", async () => {
    await opened(twinExtra, "none");
    await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    const why = within(screen.getByRole("region", { name: "Why this plant is read-only" }));
    expect(why.getByText("Generated from the twin choices twins/fan.twin.json.")).toBeInTheDocument();
    expect(
      why.getByText(
        "This plant was generated from a twin: its links are named driver_drives_driven and carry no ports.",
      ),
    ).toBeInTheDocument();
  });

  it("offers no button when no device carries the twin choices", async () => {
    const bridge = provenanceBridge(twinExtra, "none");
    const listed = bridge.workspaceOutcome;
    if (listed.ok) bridge.workspaceOutcome = success({ ...listed.value, kinds: [] });
    mountApp(bridge, hash);
    const notice = await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    expect(within(notice).queryByRole("button")).toBeNull();
  });

  it("has no accessibility violations", async () => {
    const { view } = await opened(twinExtra, "none");
    await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});

describe("a plant that is not a twin or a composition", () => {
  it("explains a hand-written domain with the binary's sentence and offers no conversion", async () => {
    await opened(
      {
        editable: {
          verdict: "read_only",
          reason: "hand_authored",
          route: { code: "copy_domain_file", sentence: "Authored as a domain file: copy the file as a new domain id." },
        },
      },
      "none",
    );
    const notice = await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    expect(within(notice).getByText("Read-only: written by hand as a domain file.")).toBeInTheDocument();
    expect(within(notice).queryByRole("button")).toBeNull();
    fireEvent.click(band().getByRole("button", { name: "Lineage" }));
    expect(screen.getByText(/copy the file as a new domain id/)).toBeInTheDocument();
  });

  it("creates an editable copy of a shipped example under a new id and opens it in the editor", async () => {
    const app = await opened(exampleExtra, "none");
    fireEvent.click(await screen.findByRole("button", { name: "Create an editable copy from this example" }));
    const form = await screen.findByRole("form", { name: "Editable copy of this example" });
    fireEvent.change(within(form).getByLabelText("Id"), { target: { value: "mini-copy" } });
    app.bridge.actions.set(
      "component.new",
      success({ composition: { ...frame, id: "mini-copy", title: "Mini plant copy" } }),
    );
    fireEvent.submit(form);
    await vi.waitFor(() => {
      expect(app.location.hash).toBe("#/edit?composition=compositions%2Fmini-copy.composition.json");
    });
    expect(app.bridge.calls.find((call) => call.name === "component.new")?.args).toMatchObject({
      example: "mini",
      id: "mini-copy",
    });
  });

  it("shows the binary's refusal of the copy at the form", async () => {
    const app = await opened(exampleExtra, "none");
    fireEvent.click(await screen.findByRole("button", { name: "Create an editable copy from this example" }));
    const form = await screen.findByRole("form", { name: "Editable copy of this example" });
    app.bridge.actions.set("component.new", failure(cannedFailure("example_unknown", 422, "no such example")));
    fireEvent.submit(form);
    expect(await within(form).findByRole("alert")).toHaveTextContent("no such example (example_unknown)");
  });
});

describe("a plant whose composition changed after it was composed", () => {
  const outOfSync = {
    composition: { file: "compositions/mini.composition.json", id: "mini", version: "0.2.0", status: "out_of_sync" },
    editable: {
      verdict: "editable",
      reason: "composition_out_of_sync",
      route: { code: "recompose", sentence: "The composition no longer produces this domain file: compose it again." },
    },
  };

  it("explains it and composes again against the plant file on disk, then reads the lineage again", async () => {
    const app = await opened(outOfSync);
    const notice = await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    expect(within(notice).getByText("This plant no longer matches its composition.")).toBeInTheDocument();
    app.bridge.addFile(miniDomainPath, { id: "mini" }, "disk-7");
    app.bridge.actions.set("domain.compose", success({ channels: 3, faults: 1, effectors: 0, digest: hash0("c") }));
    fireEvent.click(within(notice).getByRole("button", { name: "Compose again" }));
    expect(await within(notice).findByText("Composed again.")).toBeInTheDocument();
    expect(app.bridge.calls.find((call) => call.name === "domain.compose")?.args).toEqual({
      composition: "compositions/mini.composition.json",
      out: miniDomainPath,
      etag: "disk-7",
    });
    await vi.waitFor(() => {
      expect(app.bridge.calls.filter((call) => call.name === "domain.lineage")).toHaveLength(2);
    });
  });

  it("shows the binary's refusal when it cannot compose", async () => {
    const app = await opened(outOfSync);
    const notice = await screen.findByRole("region", { name: "Why this plant cannot be edited here" });
    app.bridge.actions.set("domain.compose", failure(cannedFailure("compose_failed", 422, "unknown preset")));
    fireEvent.click(within(notice).getByRole("button", { name: "Compose again" }));
    expect(await within(notice).findByRole("alert")).toHaveTextContent("unknown preset (compose_failed)");
  });
});
