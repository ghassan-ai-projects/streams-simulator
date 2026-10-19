import { act, fireEvent, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { liveWire } from "../../../test/liveFixture";
import { reviewBridge, reviewHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

const digest = `sha256:${"b".repeat(64)}`;
const other = `sha256:${"c".repeat(64)}`;

function refused(commandId: string, change: Record<string, unknown> = {}) {
  const accepted = liveWire().commands[0];
  return {
    ...accepted,
    command_id: commandId,
    receipt: "refused",
    reason: "policy_not_allowed",
    report: "unknown",
    policy_digest: digest,
    ...change,
  };
}

async function openWith(commands: readonly object[], plan: Record<string, unknown>, copied = true) {
  const bridge = reviewBridge();
  const clipboard = { copy: vi.fn(() => Promise.resolve(copied)) };
  const app = mountApp(bridge, reviewHash(), clipboard);
  await screen.findByRole("region", { name: "Run w-1" });
  act(() => {
    bridge.stream("w-1", [
      { type: "session", seq: 1, stream: "observer", at_ns: 0, mode: "connected", plan: { instances: [], ...plan } },
      { ...liveWire(), commands },
    ]);
    app.timers.advance(250);
  });
  app.stores.review.seek(10e9);
  const table = await screen.findByRole("table", { name: "Command evidence" });
  return { app, table, clipboard };
}

describe("a command refused for its policy", () => {
  it("says who refused it, shows the host's digest shortened with the full digest for assistive technology, and leaves nothing to scroll", async () => {
    const { table } = await openWith([refused("cmd-2"), refused("cmd-3", { environment: "hardware" })], {
      domain: "mini",
      policies: [other],
    });
    expect(within(table).getByText("Policy not allowed by this session", { exact: true })).toBeInTheDocument();
    expect(within(table).getByText("Policy not allowed by the gateway", { exact: true })).toBeInTheDocument();
    const [short] = within(table).getAllByText("sha256:bbbbbbbb…bbbbbbbb");
    expect(short).toHaveAttribute("title", digest);
    const copy = within(table).getAllByRole("button", { name: "Copy policy digest" })[0];
    expect(copy).toHaveAccessibleDescription(expect.stringContaining(digest));
    expect(screen.getByText("Allowed policies: 1")).toBeInTheDocument();
    const trusted = screen.getByRole("list", { name: "Digests trusted by this session" });
    expect(within(trusted).getByText("sha256:cccccccc…cccccccc")).toHaveAttribute("title", other);
    expect(trusted).toHaveTextContent(other);
    expect((await axe.run(table)).violations).toEqual([]);
  });

  it("shows no digest line on a command accepted or refused for another reason", async () => {
    const accepted = liveWire().commands[0] ?? {};
    const { table } = await openWith(
      [accepted, refused("cmd-2", { reason: "local_control", policy_digest: undefined })],
      {
        domain: "mini",
      },
    );
    expect(within(table).queryByRole("button", { name: "Copy policy digest" })).not.toBeInTheDocument();
    expect(within(table).queryByText(/Host's policy/)).not.toBeInTheDocument();
    expect(screen.getByText("Allowed policies: not recorded.")).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Digests trusted by this session" })).not.toBeInTheDocument();
  });

  it("copies the full digest only when asked, and says so for that command", async () => {
    const { table, clipboard } = await openWith([refused("cmd-2")], { domain: "mini", policies: [] });
    expect(clipboard.copy).not.toHaveBeenCalled();
    fireEvent.click(within(table).getByRole("button", { name: "Copy policy digest" }));
    expect(await within(table).findByText("Copied the policy digest of cmd-2.")).toBeInTheDocument();
    expect(clipboard.copy).toHaveBeenCalledExactlyOnceWith(digest);
    expect(
      screen.getByText("Allowed policies: none. Commands to simulated devices are refused, except safe stops."),
    ).toBeInTheDocument();
  });

  it("shows the full digest to select when the browser refuses to copy", async () => {
    const { table } = await openWith([refused("cmd-2")], { domain: "mini", policies: [] }, false);
    fireEvent.click(within(table).getByRole("button", { name: "Copy policy digest" }));
    expect(await within(table).findByText("Could not copy; select the digest instead.")).toBeInTheDocument();
    expect(within(table).getByText(digest, { selector: "code" })).toBeInTheDocument();
  });

  it("adds the digest to the plant's draft only when asked, never to the running session", async () => {
    const { app, table } = await openWith([refused("cmd-2")], { domain: "mini", policies: [] });
    const connect = app.stores.launch.connect;
    expect(connect.draftFor("mini").policies).toBe("");
    fireEvent.click(within(table).getByRole("button", { name: "Use this digest" }));
    await waitFor(() => {
      expect(connect.draftFor("mini")).toMatchObject({ policies: digest, policiesFrom: "operator" });
    });
    expect(within(table).getByText(/draft session plan of mini\. It applies to the next session/)).toBeInTheDocument();
    expect(connect.snapshot()).toEqual({ phase: "idle" });
    fireEvent.click(within(table).getByRole("button", { name: "Use this digest" }));
    expect(connect.draftFor("mini").policies).toBe(digest);
  });

  it("offers no draft to add to when the session did not record its plant", async () => {
    const { table } = await openWith([refused("cmd-2")], { policies: [] });
    expect(within(table).queryByRole("button", { name: "Use this digest" })).not.toBeInTheDocument();
    expect(within(table).getByRole("button", { name: "Copy policy digest" })).toBeInTheDocument();
  });
});
