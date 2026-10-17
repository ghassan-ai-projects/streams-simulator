import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { reviewState as state } from "../../../test/reviewFixture";
import { report } from "../../../test/runFixture";
import { Status } from "./RunStatus";

const parsed = report;
const now = () => 0;

describe("run status", () => {
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
});
