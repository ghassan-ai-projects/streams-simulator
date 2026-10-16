import { useReactFlow, useStore } from "@xyflow/react";
import { cutSide } from "../../../model/graph";
import type { CutSide } from "../../../model/graph";
import type { FlowNode } from "./flow";
import styles from "./graph.module.css";

function useCutSide(): CutSide {
  const flow = useReactFlow<FlowNode>();
  const [x, y, zoom] = useStore((state) => state.transform);
  const width = useStore((state) => state.width);
  const height = useStore((state) => state.height);
  const ready = useStore((state) => state.nodes.length > 0);
  if (!ready) return "none";
  return cutSide(flow.getNodesBounds(flow.getNodes()), { x, y, zoom }, { width, height });
}

export function PanCue() {
  const cut = useCutSide();
  if (cut === "none") return null;
  return (
    <>
      {(cut === "left" || cut === "both") && <div className={styles.panCue} data-pan-cue="left" aria-hidden="true" />}
      {(cut === "right" || cut === "both") && <div className={styles.panCue} data-pan-cue="right" aria-hidden="true" />}
    </>
  );
}
