import { useEffect, useState } from "react";
import { describe, expect, it } from "vitest";
import { renderStrict } from "./strictMode";

describe("the strict-mode harness", () => {
  it("runs an effect twice so a session that survives one mount is caught", () => {
    const runs: string[] = [];
    function Probe() {
      useEffect(() => {
        runs.push("open");
        return () => {
          runs.push("close");
        };
      }, []);
      return <p>probe</p>;
    }
    renderStrict(<Probe />);
    expect(runs).toEqual(["open", "close", "open"]);
  });

  it("re-renders a stateful component once per mount cycle", () => {
    let renders = 0;
    function Counter() {
      const [count] = useState(0);
      renders += 1;
      return <p>{count}</p>;
    }
    renderStrict(<Counter />);
    expect(renders).toBeGreaterThan(1);
  });
});
