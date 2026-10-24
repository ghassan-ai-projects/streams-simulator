import { act, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { editing } from "../../../test/editorHarness";
import { failure, success } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";

const digest = `sha256:${"c".repeat(64)}`;
const summary = { channels: 4, faults: 2, effectors: 0, digest };

async function saved(lint: Parameters<typeof success>[0]) {
  const opened = await editing();
  opened.fake.actions.set("domain.compose", success(summary));
  opened.fake.actions.set("domain.lint", success(lint));
  opened.editor.edit((doc) => ({ ...doc, title: "Changed" }));
  await act(async () => {
    await opened.editor.save();
  });
  return opened;
}

function report() {
  return within(screen.getByRole("region", { name: "Last save" }));
}

describe("what a save reports", () => {
  it("shows nothing before the first save", async () => {
    await editing();
    expect(screen.queryByRole("region", { name: "Last save" })).toBeNull();
  });

  it("shows the composition, the digest, a plausibility that passed, and that the interface does not apply", async () => {
    await saved({ errors: 0, warnings: 0, findings: [] });
    expect(await report().findByText("Readings stay plausible in a nominal day")).toBeInTheDocument();
    expect(report().getByText("Composed 4 channels, 2 faults, 0 effectors.")).toBeInTheDocument();
    expect(report().getByTitle(digest)).toHaveTextContent("sha256:cccccccc…cccccccc");
    expect(report().getByText("Interface:")).toBeInTheDocument();
    expect(report().getByText(/^Not applicable: a composed plant has no device interface/)).toBeInTheDocument();
  });

  it("lists the warnings the plausibility check found, word for word", async () => {
    await saved({
      errors: 0,
      warnings: 1,
      findings: [{ severity: "warning", entity: "CH-1", detail: "the reading is flat" }],
    });
    expect(await report().findByText("Plausible, with 1 warning")).toBeInTheDocument();
    expect(report().getByText("warning: CH-1: the reading is flat")).toBeInTheDocument();
  });

  it("lists the issues it found, and says the save still composed", async () => {
    await saved({ errors: 2, warnings: 0, findings: [{ severity: "error", entity: "CH-1", detail: "out of range" }] });
    expect(await report().findByText("2 issues")).toBeInTheDocument();
    expect(report().getByText("error: CH-1: out of range")).toBeInTheDocument();
    expect(report().getByText("Composed 4 channels, 2 faults, 0 effectors.")).toBeInTheDocument();
  });

  it("says when the check could not be made, with the binary's message", async () => {
    const opened = await editing();
    opened.fake.actions.set("domain.compose", success(summary));
    opened.fake.actions.set("domain.lint", failure(cannedFailure("domain_failed", 500, "lint broke")));
    opened.editor.edit((doc) => ({ ...doc, title: "Changed" }));
    await act(async () => {
      await opened.editor.save();
    });
    expect(await report().findByText("Could not be checked: lint broke (domain_failed)")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { view } = await saved({
      errors: 0,
      warnings: 1,
      findings: [{ severity: "warning", entity: "CH-1", detail: "flat" }],
    });
    await report().findByText("Plausible, with 1 warning");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
