import type { Edge, EdgeProps } from "@xyflow/react";
import { BaseEdge } from "@xyflow/react";
import type { Point } from "../../../model/graph";
import { labelPoint, wirePath } from "../../../model/graph";

export type WireEdgeType = Edge<{ readonly points: readonly Point[] }, "wire">;

export function WireEdge({ id, data, markerEnd, label }: EdgeProps<WireEdgeType>) {
  const points = data?.points ?? [];
  const middle = labelPoint(points);
  const marker = markerEnd === undefined ? {} : { markerEnd };
  return <BaseEdge id={id} path={wirePath(points)} {...marker} label={label} labelX={middle.x} labelY={middle.y} />;
}
