import type { Lineage, LineagePlant } from "./lineageTypes";

type PlantAction =
  | { readonly kind: "edit"; readonly composition: string }
  | { readonly kind: "recompose"; readonly composition: string; readonly domain: string }
  | { readonly kind: "behaviour"; readonly device: string }
  | { readonly kind: "example"; readonly example: string }
  | { readonly kind: "none" };

export interface PlantRoute {
  readonly headline: string;
  readonly sentence: string;
  readonly origin: string;
  readonly action: PlantAction;
  readonly label: string;
}

export function plantOf(lineage: Lineage | null, domainFile: string): LineagePlant | null {
  return lineage?.plants.find((plant) => plant.file === domainFile) ?? null;
}

function stemOfFile(path: string): string {
  return (path.split("/").at(-1) ?? "").replace(/\.(twin|kind)\.json$/, "");
}

function leadDevice(plant: LineagePlant, devices: readonly string[]): string | null {
  const stem = stemOfFile(plant.twin?.file ?? "");
  return devices.find((path) => stemOfFile(path) === stem) ?? null;
}

const twinOrigin = "This plant was generated from a twin: its links are named driver_drives_driven and carry no ports.";

const headlines: Readonly<Record<LineagePlant["editable"]["reason"], string>> = {
  composition: "",
  composition_out_of_sync: "This plant no longer matches its composition.",
  twin: "Read-only: this plant is generated from a twin.",
  twin_out_of_sync: "Read-only: this plant no longer matches the twin it was generated from.",
  example: "Read-only: a shipped example.",
  hand_authored: "Read-only: written by hand as a domain file.",
};

function actionOf(plant: LineagePlant, devices: readonly string[]): { action: PlantAction; label: string } {
  const code = plant.editable.route.code;
  const lead = leadDevice(plant, devices);
  if (code === "recompose" && plant.composition !== null) {
    return {
      action: { kind: "recompose", composition: plant.composition.file, domain: plant.file },
      label: "Compose again",
    };
  }
  if ((code === "open_device_behavior" || code === "rescaffold_twin") && lead !== null) {
    return { action: { kind: "behaviour", device: lead }, label: "Open the twin's Behaviour step" };
  }
  if (code === "create_from_example") {
    return { action: { kind: "example", example: plant.id }, label: "Create an editable copy from this example" };
  }
  return { action: { kind: "none" }, label: "" };
}

export function plantRoute(plant: LineagePlant | null, devices: readonly string[]): PlantRoute | null {
  if (plant === null || plant.editable.reason === "composition") return null;
  const { action, label } = actionOf(plant, devices);
  return {
    headline: headlines[plant.editable.reason],
    sentence: plant.editable.route.sentence,
    origin: plant.twin === null ? "" : twinOrigin,
    action,
    label,
  };
}
