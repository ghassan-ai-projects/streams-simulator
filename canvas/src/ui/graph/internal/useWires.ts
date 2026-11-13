import { useMemo, useState } from "react";
import { canGroupWires, openingWires } from "../../../model/graph";
import type { Wires } from "../../../model/graph";
import type { Plant } from "../../../model/plant";

export interface WiresControl {
  readonly value: Wires;
  readonly onChange: (wires: Wires) => void;
}

interface Choice {
  readonly plant: string;
  readonly wires: Wires;
}

interface Wiring {
  readonly wires: Wires;
  readonly control: WiresControl | null;
}

export function useWires(plant: Plant | null, opening: "follow" | "keep" = "follow"): Wiring {
  const [choice, setChoice] = useState<Choice | null>(null);
  const [first] = useState(() => (plant === null ? null : openingWires(plant)));
  return useMemo(() => {
    if (plant === null || !canGroupWires(plant)) return { wires: "each", control: null };
    const opened = opening === "keep" && first !== null ? first : openingWires(plant);
    const wires = choice?.plant === plant.id ? choice.wires : opened;
    const onChange = (next: Wires) => {
      setChoice({ plant: plant.id, wires: next });
    };
    return { wires, control: { value: wires, onChange } };
  }, [plant, choice, opening, first]);
}
