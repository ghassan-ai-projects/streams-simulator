import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { cannedFailure } from "../../../test/fakeBridge";
import { launchHarness } from "../../../test/launchFixture";
import { failure, success } from "../../../test/outcome";
import { domain } from "../../../test/plantFixture";
import { a, b, checkThePlan, checkedWith, field, file } from "../../../test/policyHarness";
import { check, fill, renderSession, sessionBridge } from "../../../test/sessionHarness";
import { LaunchPanel } from "./LaunchPanel";

function formOf(): HTMLElement {
  return screen.getByRole("form", { name: "Session plan" });
}

describe("the result of a check", () => {
  it("keeps the form's scroll area mounted across checks and moves focus to the result", async () => {
    const fake = checkedWith([a]);
    renderSession(fake);
    const scroller = formOf().firstElementChild;
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    const result = await screen.findByRole("region", { name: "Plan check result" });
    expect(result).toHaveFocus();
    expect(formOf().firstElementChild).toBe(scroller);
    expect(result).toContainElement(screen.getByRole("button", { name: "Save as default" }));
    fill("Host socket", "/tmp/other.sock");
    screen.getByLabelText("Host socket").focus();
    expect(screen.getByRole("region", { name: "Plan check result" })).not.toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    await waitFor(() => {
      expect(screen.getByRole("region", { name: "Plan check result" })).toHaveFocus();
    });
    expect(formOf().firstElementChild).toBe(scroller);
  });

  it("says the save came out next to the button it belongs to", async () => {
    const fake = checkedWith([a]);
    renderSession(fake);
    await checkThePlan();
    const save = screen.getByRole("button", { name: "Save as default" });
    fireEvent.click(save);
    const status = await screen.findByText("Saved 1 policy digest as the workspace default.");
    expect(status).toHaveAttribute("role", "status");
    expect(save.nextElementSibling).toBe(status);
  });

  it("forgets the save result when the list is edited or the plan is checked again", async () => {
    const fake = checkedWith([]);
    renderSession(fake);
    await checkThePlan();
    fireEvent.click(screen.getByRole("button", { name: "Save as default" }));
    expect(await screen.findByText("The workspace default is now empty.")).toBeInTheDocument();
    fill("Allowed policy digests", a);
    expect(screen.queryByText("The workspace default is now empty.")).not.toBeInTheDocument();
    fake.actions.set("session.check", success({ ...check, policies: [a] }));
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    await screen.findByText("Allowed policies: 1");
    expect(screen.queryByText("The workspace default is now empty.")).not.toBeInTheDocument();
    expect(screen.queryByText(/Saved \d policy digest/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save as default" })).toBeInTheDocument();
  });

  it("says fix the marked field, without the binary's raw lines, when the field carries every issue", async () => {
    const fake = sessionBridge();
    const issues = [{ pointer: "/policies", message: "array longer than maxItems 8", keyword: "maxItems" }];
    fake.actions.set("session.check", failure({ ...cannedFailure("plantwire_failed", 422, "refused"), issues }));
    renderSession(fake);
    fill("Allowed policy digests", a);
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("streamsim refused the plan: fix the marked field and check it again.");
    expect(within(alert).queryByRole("list", { name: "Problems in the plan" })).not.toBeInTheDocument();
    expect(within(alert).getByText(/array longer than maxItems 8/)).not.toBeVisible();
    expect(field()).toHaveAccessibleDescription(/List at most 8 policy digests\./);
  });

  it("keeps the other issues of the plan listed beside the marked field", async () => {
    const fake = sessionBridge();
    const issues = [
      { pointer: "/policies", message: "array longer than maxItems 8", keyword: "maxItems" },
      { pointer: "/host_socket", message: "not absolute", keyword: "socket_not_absolute" },
    ];
    fake.actions.set("session.check", failure({ ...cannedFailure("plantwire_failed", 422, "refused"), issues }));
    renderSession(fake);
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("fix what is marked below");
    const listed = within(alert).getByRole("list", { name: "Problems in the plan" });
    expect(within(listed).getAllByRole("listitem")).toHaveLength(1);
  });
});

describe("what starting would trust", () => {
  it("names the digests of the workspace default above Start, and a link to the field", async () => {
    const fake = sessionBridge();
    fake.addFile(file, { policies: [a, b] });
    renderSession(fake);
    const words = await screen.findByText(/^Trusting 2 policy digests from the workspace default: /);
    expect(words).toHaveTextContent("sha256:aaaaaaaa…aaaaaaaa, sha256:bbbbbbbb…bbbbbbbb");
    expect(screen.getByRole("button", { name: "Start the session" })).toHaveAccessibleDescription(
      /Trusting 2 policy digests/,
    );
    fireEvent.click(screen.getByRole("button", { name: "Show the allowed policy digests" }));
    expect(field()).toHaveFocus();
  });

  it("says the digests were entered once the operator edits the list, and says nothing for an empty list", async () => {
    const fake = sessionBridge();
    fake.addFile(file, { policies: [a, b] });
    renderSession(fake);
    await screen.findByText(/^Trusting 2 policy digests from the workspace default/);
    fill("Allowed policy digests", a);
    expect(screen.getByText("Trusting 1 policy digest you entered: sha256:aaaaaaaa…aaaaaaaa")).toBeInTheDocument();
    fill("Allowed policy digests", "");
    expect(screen.queryByText(/^Trusting/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Show the allowed policy digests" })).not.toBeInTheDocument();
  });
});

describe("setting up another run", () => {
  it("clears the sockets and keeps the list with a note saying where it came from, in the same mode", async () => {
    const fake = sessionBridge();
    fake.addFile(file, { policies: [a] });
    const { launch } = launchHarness(fake);
    const open = () =>
      render(
        <LaunchPanel launch={launch} plant={parsePlant(domain)} path="p" onClose={vi.fn()} onFinished={vi.fn()} />,
      );
    const first = open();
    fireEvent.click(screen.getByRole("radio", { name: /^Connected session/ }));
    await waitFor(() => {
      expect(field()).toHaveValue(a);
    });
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    await screen.findByText("Connected session · run c-1");
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    fireEvent.click(await screen.findByRole("button", { name: "Set up another run" }));
    expect(screen.getByLabelText("Host socket")).toHaveValue("");
    expect(screen.getByText("Kept from the previous run: 1 digest.")).toBeInTheDocument();
    expect(screen.queryByText("Pre-filled from the workspace default.")).not.toBeInTheDocument();
    first.unmount();
    open();
    expect(screen.getByRole("radio", { name: /^Connected session/ })).toBeChecked();
    expect(field()).toHaveValue(a);
  });

  it("opens the panel in the connected mode with the list when the review set up the next session", () => {
    const fake = sessionBridge();
    const { launch } = launchHarness(fake);
    launch.connect.setUpNext("mini", [b, a]);
    render(<LaunchPanel launch={launch} plant={parsePlant(domain)} path="p" onClose={vi.fn()} onFinished={vi.fn()} />);
    expect(screen.getByRole("radio", { name: /^Connected session/ })).toBeChecked();
    expect(field()).toHaveValue(`${b}\n${a}`);
    expect(screen.getByLabelText("Host socket")).toHaveValue("");
    expect(screen.getByText("Kept from the previous run: 2 digests.")).toBeInTheDocument();
  });

  it("opens the session plan with fresh settings and says the plan could not be recovered, until it is left", () => {
    const fake = sessionBridge();
    const { launch } = launchHarness(fake);
    launch.connect.setUpFresh("mini");
    const open = () =>
      render(
        <LaunchPanel launch={launch} plant={parsePlant(domain)} path="p" onClose={vi.fn()} onFinished={vi.fn()} />,
      );
    const first = open();
    expect(screen.getByRole("radio", { name: /^Connected session/ })).toBeChecked();
    expect(screen.getByLabelText("Host socket")).toHaveValue("");
    const note = screen.getByRole("note");
    expect(note).toHaveTextContent(
      "The plan of this run could not be recovered because it did not finish; these are fresh settings.",
    );
    expect(formOf().firstElementChild).toContainElement(note);
    first.unmount();
    open();
    expect(screen.queryByRole("note")).not.toBeInTheDocument();
  });
});
