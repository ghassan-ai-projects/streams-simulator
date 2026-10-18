import { originLabels } from "../../shared/origin";
import { counted } from "../../shared/words";
import type { BindingNote, PlanIssue, PlannedInstance, SessionCheck } from "./session";

const notBoundWords = "Not bound: refused · not announced to the host";

const bindingNextStep = "Stop the session, fix the board or device, then start it again.";

function realWords(planned: PlannedInstance, binding: BindingNote | null): string {
  if (binding !== null && !binding.matched) return `${planned.instance} · ${notBoundWords}`;
  return `${planned.instance} · ${originLabels.real} · device ${planned.device} on board ${planned.board}`;
}

export function plannedWords(planned: PlannedInstance, binding: BindingNote | null = null): string {
  if (planned.role === "simulated") return `${planned.instance} · ${originLabels.simulated}`;
  return realWords(planned, binding);
}

export function checkWords(check: SessionCheck): string {
  const real = check.instances.filter((planned) => planned.role === "real").length;
  const simulated = counted(check.instances.length - real, "simulated instance");
  if (real === 0) return `The plan holds: ${simulated}${check.mixed ? " beside the gateway's boards" : ""}.`;
  return `The plan holds: ${counted(real, "real board")} and ${simulated}.`;
}

export function bindingWords(note: BindingNote): string {
  if (note.matched) return "The gateway announced every real board of the plan with the interface of its kind file.";
  return `The gateway was refused (${note.refusal}), so its real boards were not bound. ${bindingNextStep}`;
}

const keywordWords: Readonly<Record<string, string>> = {
  socket_not_absolute: "Give the full path of the socket, starting with /.",
  socket_control_character: "The socket path holds a control character: type it again.",
  socket_parent_component: "Write the socket path without a .. part.",
  gateway_required: "A real board needs a gateway: give the gateway socket.",
  same_socket: "The gateway socket must differ from the host socket.",
  unknown_instance: "The plant has no such instance.",
  no_kind_file: "This instance's kind has no kind file in the workspace yet: capture its board first.",
  instance_bound_twice: "This instance is already bound to another real board.",
  device_bound_twice: "This device is already bound to another instance.",
  board_missing: "The gateway announces no such board.",
  device_missing: "The board holds no such device.",
  kind_missing: "The device announces no interface for this instance's kind.",
  kind_mismatch: "The device's interface differs from this instance's kind file.",
  device_kind_mismatch: "The device announced a different kind than the plan binds it to.",
  inventory_mismatch: "The gateway's boards differ from the plan.",
  environment_mismatch: "The gateway is not the kind of gateway the plan expects.",
};

const memberWords: Readonly<Record<string, string>> = {
  seed: "The seed is a whole number, such as 1.",
  host_socket: "Give the host socket: the full path the plant wire host listens on.",
  gateway_socket: "Give the gateway socket as a full path, or leave it empty.",
  board: "Give the board's name as the gateway announces it.",
  device: "Give the device's name on that board.",
  instance: "Pick the instance this real board stands for.",
};

function memberOf(issue: PlanIssue): string {
  return issue.member ?? issue.pointer.split("/").at(-1) ?? "";
}

export function issueWords(issue: PlanIssue): string {
  return keywordWords[issue.keyword ?? ""] ?? memberWords[memberOf(issue)] ?? issue.message;
}

export function issueDetail(issue: PlanIssue): string {
  const keyword = issue.keyword === undefined ? "" : `${issue.keyword} · `;
  return `${keyword}${issue.pointer === "" ? "/" : issue.pointer} · ${issue.message}`;
}

export const sessionStartWords =
  "Start links the plant to the plant wire host, and through the gateway to its real boards, and runs until you stop it. " +
  "Stop ends the session and saves it as a run you can review.";

export const staleCheckWords = "The plan changed since this check: check it again before you start.";
