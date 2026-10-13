import { render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import type { ReviewState } from "../../app/reviewStore";
import { parseReport } from "../../model/report";
import { report } from "../../test/runFixture";
import { Banners, Status } from "./RunStatus";

const parsed = parseReport(report);

function state(overrides: Partial<ReviewState>): ReviewState {
  return {
    worldId: "w-1",
    report: { phase: "ready", value: parsed },
    verified: { phase: "ready", value: { matches: true, domainChanged: false, detail: "" } },
    playback: { phase: "ready", records: parsed.counts.delivered ?? 0, adapter: "native-jsonl" },
    positionNs: 0,
    selected: null,
    runDomain: { phase: "loading" },
    tab: "report",
    openedAt: 0,
    ...overrides,
  };
}

describe("run status", () => {
  it("badges a divergent replay Not verified with no warning banner", () => {
    const divergent = state({
      verified: { phase: "ready", value: { matches: false, domainChanged: false, detail: "trace differs" } },
    });
    render(
      <>
        <Status state={divergent} />
        <Banners state={divergent} />
      </>,
    );
    expect(screen.getByText("Not verified")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("says when the plant file changed after the run, as information", async () => {
    const changed = state({
      verified: { phase: "ready", value: { matches: true, domainChanged: true, detail: "" } },
    });
    const view = render(<Banners state={changed} />);
    expect(screen.getByRole("status")).toHaveTextContent("The plant file changed after this run");
    expect(screen.queryByRole("alert")).toBeNull();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("warns in plain words when verify could not run, with the binary's message under Details", () => {
    const failed = state({
      verified: { phase: "failed", failure: { status: 422, code: "x", message: "run.json missing", issues: [] } },
    });
    render(<Banners state={failed} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Not checked: streamsim verify could not replay this run.");
    expect(screen.getByText("run.json missing")).toBeInTheDocument();
  });

  it("warns when playback holds fewer observations than the report delivered", () => {
    render(<Banners state={state({ playback: { phase: "ready", records: 1, adapter: "native-jsonl" } })} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Playback has 1 observations");
  });
});
