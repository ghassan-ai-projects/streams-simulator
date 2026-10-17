import { useState } from "react";
import { firstOf, hallCount } from "../../../model/graph";
import type { PlantView } from "../../../model/graph";

export function useSelectionMoves(view: PlantView, selected: string | null, setActive: (id: string) => void) {
  const [seen, setSeen] = useState(selected);
  if (seen !== selected) {
    setSeen(selected);
    if (selected !== null && view.nodes.some((entry) => entry.id === selected)) setActive(selected);
  }
}

export function useViewActive(view: PlantView, viewKey: string) {
  const [active, setActive] = useState<string | null>(null);
  const [shownKey, setShownKey] = useState(viewKey);
  const [entered, setEntered] = useState("");
  if (shownKey !== viewKey) {
    const back = view.nodes.some((entry) => entry.id === shownKey);
    setShownKey(viewKey);
    setActive(back ? shownKey : firstOf(view.layout.nodes));
    setEntered(back ? "" : `${viewKey}, ${hallCount(view, viewKey)}. `);
  }
  const current = active !== null && view.nodes.some((entry) => entry.id === active) ? active : null;
  const move = (id: string | null) => {
    setEntered("");
    setActive(id);
  };
  return { active: current, setActive: move, entered };
}
