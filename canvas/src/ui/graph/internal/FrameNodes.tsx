import type { NodeProps } from "@xyflow/react";
import { countWords, splitName } from "../../../model/graph";
import type { FrameTitleNodeType } from "./flow";
import frames from "./frames.module.css";

export function FrameNode() {
  return <div className={frames.frame} />;
}

export function FrameTitleNode({ data }: NodeProps<FrameTitleNodeType>) {
  const count = countWords(data.count);
  const { head, tail } = splitName(data.title);
  return (
    <div className={frames.frameTitleNode}>
      <div className={frames.frameTitle} data-part="frame-title" title={`${data.title} · ${count}`}>
        <span className={frames.frameName} data-part="frame-name">
          <span className={frames.nameHead}>{head}</span>
          <span className={frames.nameTail} data-part="name-tail">
            {tail}
          </span>
        </span>
        <span className={frames.frameCount} data-part="frame-count">
          {count}
        </span>
      </div>
    </div>
  );
}
