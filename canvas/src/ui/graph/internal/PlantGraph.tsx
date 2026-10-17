import { firstOf, openingFloor } from "../../../model/graph";
import { FlowCanvas } from "./FlowCanvas";
import type { PlantGraphProps } from "./types";
import { useGraphKeys } from "./useGraphKeys";
import { useFitted, useFollowSelection } from "./viewportFit";
import styles from "./graph.module.css";
import "./vendor.module.css";

export function PlantGraph(props: PlantGraphProps) {
  const keys = useGraphKeys(props);
  const anchor = props.anchor ?? firstOf(props.view.layout.nodes);
  useFitted(props.fitKey ?? props.viewKey, openingFloor, anchor);
  useFollowSelection(keys.active ?? props.selected, props.focusZoom ?? 0);
  return (
    <div className={styles.graph} onPointerDown={keys.onPointerDown}>
      <p id="graph-keys" className="sr-only">
        Arrow keys move between instances, Space selects, Enter selects an instance or opens a hall, Escape clears the
        selection, question mark lists every key. The outline lists every instance.
      </p>
      <p id="graph-said" className="sr-only" aria-live="polite">
        {keys.announcement}
      </p>
      <FlowCanvas
        {...props}
        keys={keys}
        active={keys.active}
        onSelect={(id) => {
          keys.point(id);
          props.onSelect(id);
        }}
      />
    </div>
  );
}
