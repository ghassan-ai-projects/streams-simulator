import { useState } from "react";
import type { BaseChoice, MakeMode, PlantMaker } from "../../../app/plant";
import { idProblemOf } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import { suggestedId } from "../../../model/workspace";

export function useCreate(
  { maker, navigation }: { readonly maker: PlantMaker; readonly navigation: NavigationStore },
  reportId: (id: string, message: string) => void,
) {
  const [failure, setFailure] = useState("");
  const create = async (base: BaseChoice, id: string, title: string, mode: MakeMode) => {
    const made = await maker.create(base.source, id, title, mode);
    if (made.ok) {
      navigation.go({ screen: "edit", composition: made.value.composition });
      return;
    }
    reportId(id, idProblemOf(made.failure));
    setFailure(made.failure.diagnostic.message);
  };
  return { failure, create };
}

export function useNaming(taken: readonly string[]) {
  const [name, setName] = useState("");
  const [typed, setTyped] = useState<string | null>(null);
  const id = typed ?? (name.trim() === "" ? "" : suggestedId(name, taken));
  const followAgain = () => {
    if (typed === "") setTyped(null);
  };
  return { name, setName, id, setId: setTyped, followAgain };
}

export function useIdProblem(
  maker: PlantMaker,
  base: BaseChoice | null,
  name: string,
  id: string,
  followAgain: () => void,
) {
  const [found, setFound] = useState<{ readonly id: string; readonly message: string } | null>(null);
  const reportId = (value: string, message: string) => {
    setFound({ id: value, message });
  };
  const check = async () => {
    const value = id.trim();
    if (value === "" || base === null) return;
    reportId(value, await maker.idProblem(base.source, value, name.trim() === "" ? value : name.trim()));
  };
  const leave = () => {
    followAgain();
    void check();
  };
  return { problem: found !== null && found.id === id.trim() ? found.message : "", leave, reportId };
}
