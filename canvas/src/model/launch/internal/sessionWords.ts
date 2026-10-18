import { counted } from "../../shared/words";
import type { BindingNote, InstanceRole, PlannedInstance, SessionCheck } from "./session";

export const roleWords: Readonly<Record<InstanceRole, string>> = { real: "Real board", simulated: "Simulated" };

export function plannedWords(planned: PlannedInstance): string {
  if (planned.role === "simulated") return `${planned.instance} · ${roleWords.simulated}`;
  return `${planned.instance} · ${roleWords.real} · device ${planned.device} on board ${planned.board}`;
}

export function checkWords(check: SessionCheck): string {
  const real = check.instances.filter((planned) => planned.role === "real").length;
  const simulated = counted(check.instances.length - real, "simulated instance");
  if (real === 0) return `The plan holds: ${simulated}${check.mixed ? " beside the gateway's boards" : ""}.`;
  return `The plan holds: ${counted(real, "real board")} and ${simulated}.`;
}

export function bindingWords(note: BindingNote): string {
  if (note.matched) return "The gateway announced every real board of the plan with the interface of its kind file.";
  const why = note.reason === "" ? "" : `: ${note.reason}`;
  return `The gateway was refused (${note.refusal})${why}.`;
}

export const sessionStartWords =
  "Start links the plant to the plant wire host, and through the gateway to its real boards, and runs until you stop it. " +
  "Stop ends the session and saves it as a run you can review.";
