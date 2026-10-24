import { useEffect } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { FaultCategories, PerturbationKind } from "../../../model/launch";
import { useStore } from "../../primitives";

const none: readonly PerturbationKind[] = [];
const noCategories: FaultCategories = {};

export function usePlanCatalog(launch: LaunchStore, domain: string) {
  const state = useStore(launch.catalog);
  useEffect(() => {
    launch.catalog.load(domain);
  }, [launch, domain]);
  const { perturbations, faultCategories } = state;
  return {
    perturbations,
    kinds: perturbations.status === "ready" ? perturbations.value : none,
    categories: faultCategories.status === "ready" ? faultCategories.value : noCategories,
  };
}
