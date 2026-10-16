import { render, screen } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { beforeAll, describe, expect, it } from "vitest";
import type { Dump } from "../../test/wireAudit";
import { viewOf } from "../../test/wireAudit";
import { PlantGraph } from "./PlantGraph";

class NoopResizeObserver {
  observe(): void {
    return undefined;
  }
  unobserve(): void {
    return undefined;
  }
  disconnect(): void {
    return undefined;
  }
}

beforeAll(() => {
  globalThis.ResizeObserver = NoopResizeObserver;
});

const oneNode: Dump = {
  nodes: [{ id: "CH-1", x: 0, y: 0, width: 232, height: 152, ins: ["heat_in"], outs: ["heat_out"] }],
  frames: [],
  edges: [],
};

function graph(onConnect?: (from: string, to: string) => void) {
  return render(
    <ReactFlowProvider>
      <PlantGraph
        view={viewOf(oneNode)}
        viewKey="v"
        selected={null}
        onSelect={() => undefined}
        {...(onConnect === undefined ? {} : { onConnect })}
      />
    </ReactFlowProvider>,
  );
}

describe("the plant graph's own description", () => {
  it("promises selection only where nothing can be deleted", () => {
    graph();
    expect(screen.getByText("Press enter or space to select a node. Escape clears the selection.")).toBeInTheDocument();
    expect(screen.queryByText(/Press delete to remove it/)).not.toBeInTheDocument();
  });

  it("keeps the delete wording in the editor graph, where an instance can be removed", () => {
    graph(() => undefined);
    expect(
      screen.getByText("Press enter or space to select a node. Press delete to remove it and escape to cancel."),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Press enter or space to select a node. Escape clears the selection."),
    ).not.toBeInTheDocument();
  });
});
