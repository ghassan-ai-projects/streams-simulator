import type { Connection } from "@xyflow/react";
import { nodeTypes } from "./nodeTypes";
import type { PlantGraphProps } from "./types";
import { lowestZoom } from "./viewportFit";
import { WireEdge } from "./WireEdge";

const edgeTypes = { wire: WireEdge };

export const graphNames = {
  "aria-roledescription": "plant graph",
  "aria-label": "Plant graph",
  "aria-describedby": "graph-keys",
  tabIndex: 0,
} as const;

export const graphWords = {
  select: {
    "node.a11yDescription.default": "Press enter or space to select a node. Escape clears the selection.",
    "node.a11yDescription.keyboardDisabled": "Press enter or space to select a node. Escape clears the selection.",
    "edge.a11yDescription.default": "Press enter or space to select an edge. Escape clears the selection.",
  },
  edit: {
    "node.a11yDescription.default":
      "Press enter or space to select a node. Press delete to remove it and escape to cancel.",
    "node.a11yDescription.keyboardDisabled":
      "Press enter or space to select a node. Press delete to remove it and escape to cancel.",
    "edge.a11yDescription.default":
      "Press enter or space to select an edge. You can then press delete to remove it or escape to cancel.",
  },
} as const;

export const fixedProps = {
  nodeTypes,
  edgeTypes,
  nodesDraggable: false,
  nodesFocusable: false,
  edgesFocusable: false,
  disableKeyboardA11y: true,
  zoomOnDoubleClick: false,
  onlyRenderVisibleElements: true,
  minZoom: lowestZoom,
  proOptions: { hideAttribution: true },
} as const;

export function connection(onConnect: PlantGraphProps["onConnect"]) {
  if (onConnect === undefined) return { nodesConnectable: false };
  return {
    nodesConnectable: true,
    onConnect: (made: Connection) => {
      if (made.sourceHandle && made.targetHandle) {
        onConnect(`${made.source}.${made.sourceHandle}`, `${made.target}.${made.targetHandle}`);
      }
    },
  };
}
