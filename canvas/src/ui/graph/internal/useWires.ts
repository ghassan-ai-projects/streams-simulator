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

export function useWires(plant: Plant | null): { readonly wires: Wires; readonly control: WiresControl | null } {
  const [choice, setChoice] = useState<Choice | null>(null);
  return useMemo(() => {
    if (plant === null || !canGroupWires(plant)) return { wires: "each", control: null };
    const wires = choice?.plant === plant.id ? choice.wires : openingWires(plant);
    const onChange = (next: Wires) => {
      setChoice({ plant: plant.id, wires: next });
    };
    return { wires, control: { value: wires, onChange } };
  }, [plant, choice]);
}
