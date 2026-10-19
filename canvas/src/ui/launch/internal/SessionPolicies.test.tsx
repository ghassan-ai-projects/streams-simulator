import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { check, fill, renderSession, sessionBridge } from "../../../test/sessionHarness";

const a = `sha256:${"a".repeat(64)}`;
const b = `sha256:${"b".repeat(64)}`;
const file = "settings/default.policies.json";
const field = () => screen.getByLabelText("Allowed policy digests");

function checkedWith(policies: readonly string[]) {
  const fake = sessionBridge();
  fake.persistWrites();
  fake.actions.set("session.check", success({ ...check, policies }));
  return fake;
}

async function checkThePlan() {
  fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
  await screen.findByRole("region", { name: "Allowed policies in effect" });
}

describe("the allowed policy digests of a session plan", () => {
  it("offers an empty field with what it means, and sends the typed lines in the plan", async () => {
    const fake = sessionBridge();
    const { view } = renderSession(fake);
    expect(field()).toHaveValue("");
    expect(field()).toHaveAccessibleDescription(/One per line, at most 8\..*safe stops always run\./);
    expect(screen.queryByRole("button", { name: "Clear the list" })).not.toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
    fill("Allowed policy digests", ` ${b} \n\n${a}\n`);
    await checkThePlan();
    const plan = JSON.parse(String(fake.calls.find((call) => call.name === "session.check")?.args.content)) as unknown;
    expect(plan).toMatchObject({ policies: [b, a] });
  });

  it("pre-fills a new draft from the workspace default and says so, until the operator edits or clears it", async () => {
    const fake = sessionBridge();
    fake.addFile(file, { policies: [a] });
    renderSession(fake);
    await waitFor(() => {
      expect(field()).toHaveValue(a);
    });
    expect(field()).toHaveAccessibleDescription(/Pre-filled from the workspace default\./);
    fill("Allowed policy digests", `${a}\n${b}`);
    expect(field()).not.toHaveAccessibleDescription(/Pre-filled/);
    expect(screen.queryByText("Pre-filled from the workspace default.")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Clear the list" }));
    expect(field()).toHaveValue("");
    expect(field()).toHaveFocus();
  });

  it("pre-fills nothing, and says why, when the default cannot be read", async () => {
    const fake = sessionBridge();
    fake.addFile(file, "{ not json");
    renderSession(fake);
    expect(await screen.findByText("The workspace default could not be read.")).toBeInTheDocument();
    expect(field()).toHaveValue("");
    expect(field()).toHaveAccessibleDescription(/could not be read/);
  });

  it("puts the binary's issues on the field by digest, ties them to it, and focuses it", async () => {
    const fake = sessionBridge();
    const issues = [
      { pointer: "/policies/1", message: "twice", keyword: "policy_listed_twice" },
      { pointer: "/policies", message: "too many", keyword: "maxItems" },
    ];
    fake.actions.set("session.check", failure({ ...cannedFailure("plantwire_failed", 422, "refused"), issues }));
    const { view } = renderSession(fake);
    fill("Host socket", "/tmp/h.sock");
    fill("Allowed policy digests", `${a}\n${a}`);
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    await screen.findByRole("alert");
    expect(field()).toHaveAttribute("aria-invalid", "true");
    expect(field()).toHaveFocus();
    expect(field()).toHaveAccessibleDescription(
      /Digest 2: This digest is already listed\. List at most 8 policy digests\./,
    );
    expect((await axe.run(view.container)).violations).toEqual([]);
    fill("Allowed policy digests", a);
    expect(field()).not.toHaveAttribute("aria-invalid");
  });
});

describe("the list in effect and the workspace default", () => {
  it("shows the list as the binary sorted it, and saves that list, not the typed one, only after a fresh passing check", async () => {
    const fake = checkedWith([a, b]);
    const { view } = renderSession(fake);
    expect(screen.queryByRole("button", { name: "Save as default" })).not.toBeInTheDocument();
    fill("Allowed policy digests", `${b}\n${a}`);
    await checkThePlan();
    const shown = screen.getByRole("region", { name: "Allowed policies in effect" });
    expect(shown).toHaveTextContent("Allowed policies: 2");
    expect(
      within(shown)
        .getAllByRole("listitem")
        .map((item) => item.textContent),
    ).toEqual([`sha256:aaaaaaaa…aaaaaaaa${a}`, `sha256:bbbbbbbb…bbbbbbbb${b}`]);
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Save as default" }));
    expect(await screen.findByText("Saved 2 policy digests as the workspace default.")).toBeInTheDocument();
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toEqual({ policies: [a, b] });
    expect(fake.writes[0]).toMatchObject({ path: file, etag: "" });
    fill("Allowed policy digests", a);
    expect(screen.queryByRole("button", { name: "Save as default" })).not.toBeInTheDocument();
    expect(screen.getByText(/The plan changed since this check/)).toBeInTheDocument();
  });

  it("says no command to a simulated device runs when the list is empty, and can save the empty list", async () => {
    const fake = checkedWith([]);
    renderSession(fake);
    await checkThePlan();
    expect(screen.getByText(/Allowed policies: none\. Commands to simulated devices are refused/)).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Digests in effect" })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Save as default" }));
    expect(await screen.findByText("The workspace default is now empty.")).toBeInTheDocument();
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toEqual({ policies: [] });
  });

  it("reports a default that changed elsewhere and saves nothing", async () => {
    const fake = checkedWith([a]);
    fake.addFile(file, { policies: [a] }, "e0");
    renderSession(fake);
    await waitFor(() => {
      expect(field()).toHaveValue(a);
    });
    await checkThePlan();
    fake.addFile(file, { policies: [b] }, "elsewhere");
    fireEvent.click(screen.getByRole("button", { name: "Save as default" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The workspace default changed elsewhere, so nothing was saved.",
    );
    expect(fake.files.get(file)).toMatchObject({ value: { etag: "elsewhere" } });
  });

  it("reports a write the bridge refused in its own words", async () => {
    const fake = checkedWith([a]);
    fake.writeOutcomes.set(file, failure(cannedFailure("bridge_failed", 500, "disk full")));
    renderSession(fake);
    await checkThePlan();
    fireEvent.click(screen.getByRole("button", { name: "Save as default" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("The workspace default was not saved: disk full");
  });

  it("keeps showing the list the running session was started with", async () => {
    const fake = checkedWith([a]);
    fake.actions.set(
      "session.start",
      success({ run: "c-1", state: "running", domain: "mini", plan: { ...check, policies: [a] } }),
    );
    renderSession(fake);
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    await screen.findByText("Connected session · run c-1");
    const shown = screen.getByRole("region", { name: "Allowed policies in effect" });
    expect(shown).toHaveTextContent("Allowed policies: 1");
    expect(shown).toHaveTextContent("sha256:aaaaaaaa…aaaaaaaa");
    expect(screen.queryByLabelText("Allowed policy digests")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Save as default" })).not.toBeInTheDocument();
  });

  it("keeps the list, and where it came from, when the operator sets up another run", async () => {
    const fake = sessionBridge();
    fake.addFile(file, { policies: [a] });
    renderSession(fake);
    await waitFor(() => {
      expect(field()).toHaveValue(a);
    });
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    await screen.findByText("Connected session · run c-1");
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    fireEvent.click(await screen.findByRole("button", { name: "Set up another run" }));
    expect(screen.getByLabelText("Host socket")).toHaveValue("");
    expect(field()).toHaveValue(a);
    expect(screen.getByText("Pre-filled from the workspace default.")).toBeInTheDocument();
  });
});
