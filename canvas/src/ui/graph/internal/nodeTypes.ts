import { memo } from "react";
import { FrameNode, FrameTitleNode } from "./FrameNodes";
import { HallNode } from "./HallNode";
import { InstanceNode } from "./InstanceNode";

export const nodeTypes = {
  instance: memo(InstanceNode),
  hall: memo(HallNode),
  frame: memo(FrameNode),
  frameTitle: memo(FrameTitleNode),
};
