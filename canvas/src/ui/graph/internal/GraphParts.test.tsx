import { render, screen } from "@testing-library/react";
import { act } from "react";
import { describe, expect, it, vi } from "vitest";
import { stubResizeObserver } from "../../../test/dom";
import { useCanvasAspect } from "./canvasAspect";
import { enterHall } from "./enterHall";
import { connection } from "./flowConfig";
import { IdText } from "./IdText";
import { Pins } from "./Pins";

describe("an instance id on the graph", () => {
  it("is written whole at full zoom", () => {
    render(<IdText label="CH-1" zoom={1} words={null} />);
    expect(screen.getByText("CH-1")).toBeInTheDocument();
  });

  it("is written whole, sized to fit, when it has room zoomed out", () => {
    render(<IdText label="CH-1" zoom={0.5} words={null} />);
    const text = screen.getByText("CH-1");
    expect(text.style.fontSize).not.toBe("");
    expect(text.querySelector("[aria-hidden='true']")).toBeNull();
  });

  it("is shortened for the eye but kept whole for screen readers when it cannot fit", () => {
    const label = "plant-hall-12/very-long-equipment-identifier-0042";
    const { container } = render(<IdText label={label} zoom={0.3} words={30} />);
    const shown = container.querySelector("[aria-hidden='true']");
    expect(shown?.textContent).toContain("…");
    expect(screen.getByText(label)).toHaveClass("sr-only");
  });
});

describe("connecting two pins", () => {
  const made = { source: "CH-1", sourceHandle: "heat_out", target: "CT-1", targetHandle: "heat_in" };

  it("is off when the graph has no way to record a connection", () => {
    expect(connection(undefined)).toEqual({ nodesConnectable: false });
  });

  it("names both ends by instance and port", () => {
    const onConnect = vi.fn();
    const config = connection(onConnect);
    if (!("onConnect" in config)) throw new Error("not connectable");
    config.onConnect({ ...made });
    expect(onConnect).toHaveBeenCalledWith("CH-1.heat_out", "CT-1.heat_in");
  });

  it("ignores a connection that is missing a pin", () => {
    const onConnect = vi.fn();
    const config = connection(onConnect);
    if (!("onConnect" in config)) throw new Error("not connectable");
    config.onConnect({ ...made, targetHandle: null });
    expect(onConnect).not.toHaveBeenCalled();
  });
});

describe("entering a hall", () => {
  it("selects the hall where the graph cannot open one", () => {
    const onSelect = vi.fn();
    enterHall("h01", onSelect);
    expect(onSelect).toHaveBeenCalledWith("h01");
  });

  it("clears the selection and opens the hall where it can", () => {
    const onSelect = vi.fn();
    const onOpen = vi.fn();
    enterHall("h01", onSelect, onOpen);
    expect(onSelect).toHaveBeenCalledWith(null);
    expect(onOpen).toHaveBeenCalledWith("h01");
  });
});

function Aspect({ box }: { readonly box: DOMRect }) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  return (
    <div
      ref={(node) => {
        if (node !== null) node.getBoundingClientRect = () => box;
        ref.current = node;
      }}
    >
      {String(aspect)}
    </div>
  );
}

describe("the canvas aspect", () => {
  const rect = (width: number, height: number) => ({ width, height }) as DOMRect;

  it("is the host's width over its height, rounded to a quarter, once it has a size", () => {
    const watchers = stubResizeObserver();
    render(<Aspect box={rect(1000, 600)} />);
    expect(screen.getByText("undefined")).toBeInTheDocument();
    act(() => {
      watchers[0]?.trigger();
    });
    expect(screen.getByText("1.75")).toBeInTheDocument();
  });

  it("stays unknown while the host has no size", () => {
    const watchers = stubResizeObserver();
    render(<Aspect box={rect(0, 0)} />);
    act(() => {
      watchers[0]?.trigger();
    });
    expect(screen.getByText("undefined")).toBeInTheDocument();
  });

  it("stops watching when the host goes away", () => {
    const watchers = stubResizeObserver();
    render(<Aspect box={rect(10, 10)} />).unmount();
    expect(watchers[0]?.disconnected).toBe(true);
  });
});

describe("an instance with no ports on the graph", () => {
  it("says its component file lists none, rather than that data is missing from the plant", () => {
    render(<Pins ins={[]} outs={[]} editable={false} zoom={1} />);
    const note = screen.getByText("no port metadata");
    expect(note).toHaveAttribute("title", "The component file lists no ports, so none can be drawn");
    expect(screen.queryByText(/no port data/)).not.toBeInTheDocument();
  });
});
