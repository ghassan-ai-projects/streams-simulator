import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { reviewState as state } from "../../../test/reviewFixture";
import { report } from "../../../test/runFixture";
import { Status } from "./RunStatus";

const parsed = report;
const now = () => 0;

describe("run status", () => {
  it("labels a connected session as one, running while it runs, with its gateway binding, never Evaluate", () => {
    const refused = { matched: false, refusal: "inventory_mismatch", reason: "", issues: [] };
    const live = state({
      connection: "running",
      binding: { atNs: 0, binding: refused },
      report: { phase: "ready", value: { ...parsed, ended: false, complete: false } },
    });
    const { unmount } = render(<Status state={live} now={now} />);
    expect(screen.getByText("Connected session · running")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("Running");
    expect(screen.getByText("Gateway refused · real boards not bound")).toHaveAttribute(
      "title",
      expect.stringContaining("The gateway was refused (inventory_mismatch)"),
    );
    expect(screen.queryByText(/Evaluate|Did not finish/)).toBeNull();
    unmount();
    render(<Status state={state({ connection: "saved" })} now={now} />);
    expect(screen.getByText("Connected session")).toBeInTheDocument();
    expect(screen.queryByText(/Gateway/)).toBeNull();
  });

  it("badges a divergent replay Not verified", () => {
    const divergent = state({
      verified: {
        phase: "ready",
        value: { matches: false, domainChanged: false, detail: "trace differs", reason: "", file: "" },
      },
    });
    render(<Status state={divergent} now={now} />);
    expect(screen.getByRole("status")).toHaveTextContent("Not verified");
    expect(document.querySelector("[title^='Not verified:']")).toHaveTextContent("Not verified");
  });

  it("says a run that did not finish in plain words, with its reason", () => {
    const unfinished = state({
      report: { phase: "ready", value: { ...parsed, complete: false, incompleteReason: "the world was destroyed" } },
    });
    render(<Status state={unfinished} now={now} />);
    expect(screen.getByRole("status")).toHaveTextContent("Did not finish");
    expect(screen.getByText("the world was destroyed")).toBeInTheDocument();
  });

  it("names a run the listing still shows as open Did not finish, never not finished", () => {
    const open = state({ report: { phase: "ready", value: { ...parsed, ended: false, complete: false } } });
    render(<Status state={open} now={now} />);
    expect(screen.getByRole("status")).toHaveTextContent("Did not finish");
    expect(screen.queryByText("not finished")).toBeNull();
  });

  it("says a run the listing shows as running is Running, never Did not finish, while it runs", () => {
    const open = { phase: "ready", value: { ...parsed, ended: false, complete: false } } as const;
    const { unmount } = render(<Status state={state({ report: open, running: true })} now={now} />);
    expect(screen.getByRole("status")).toHaveTextContent("Running");
    expect(screen.queryByText(/Did not finish/)).toBeNull();
    expect(screen.queryByText(/Connected session/)).toBeNull();
    unmount();
    render(<Status state={state({ report: open, running: false })} now={now} />);
    expect(screen.getByRole("status")).toHaveTextContent("Did not finish");
  });

  it("styles Not verified, Did not finish, Verified and Verifying as four different chips", () => {
    const divergent = {
      phase: "ready",
      value: { matches: false, domainChanged: false, detail: "", reason: "", file: "" },
    } as const;
    const incomplete = { phase: "ready", value: { ...parsed, complete: false } } as const;
    const looks = [
      state({ verified: divergent }),
      state({ report: incomplete }),
      state({ verified: { phase: "loading" } }),
      state({ verified: { phase: "ready", value: { ...divergent.value, matches: true } } }),
    ].map((review) => {
      const view = render(<Status state={review} now={now} />);
      const chip = view.container.querySelector("[data-part='status-chip']");
      view.unmount();
      return chip?.className;
    });
    expect(new Set(looks).size).toBe(4);
  });

  it("says where the operator stopped a run, out of what was requested, and nothing for one that completed", () => {
    const stopped = {
      ...parsed,
      requestedEndNs: 3_600e9,
      endNs: 1_680e9,
      termination: { reason: "stopped_by_operator", atNs: 1_680e9 },
    } as const;
    const { unmount } = render(<Status state={state({ report: { phase: "ready", value: stopped } })} now={now} />);
    expect(screen.getByText("Stopped by operator at 00:28:00 of 01:00:00")).toBeInTheDocument();
    unmount();
    const completed = { ...parsed, termination: { reason: "completed", atNs: 3_600e9 } } as const;
    render(<Status state={state({ report: { phase: "ready", value: completed } })} now={now} />);
    expect(screen.queryByText(/Stopped by operator/)).toBeNull();
  });
});
