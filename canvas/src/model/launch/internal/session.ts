export interface RealRow {
  readonly key: number;
  readonly instance: string;
  readonly board: string;
  readonly device: string;
}

type PolicySource = "operator" | "default";

export interface SessionDraft {
  readonly seed: string;
  readonly hostSocket: string;
  readonly gatewaySocket: string;
  readonly policies: string;
  readonly policiesFrom: PolicySource;
  readonly real: readonly RealRow[];
  readonly nextKey: number;
}

export type RowField = "instance" | "board" | "device";

export type PlanMember = "seed" | "host_socket" | "gateway_socket";

export interface PlanIssue {
  readonly pointer: string;
  readonly message: string;
  readonly keyword?: string;
  readonly member?: string;
}

type InstanceRole = "real" | "simulated";

export interface PlannedInstance {
  readonly instance: string;
  readonly role: InstanceRole;
  readonly kind: string;
  readonly board: string;
  readonly device: string;
}

export interface SessionCheck {
  readonly domain: string;
  readonly mixed: boolean;
  readonly instances: readonly PlannedInstance[];
  readonly policies: readonly string[];
}

export interface BindingNote {
  readonly matched: boolean;
  readonly refusal: string;
  readonly reason: string;
  readonly issues: readonly PlanIssue[];
}

export const firstSessionDraft: SessionDraft = {
  seed: "1",
  hostSocket: "",
  gatewaySocket: "",
  policies: "",
  policiesFrom: "operator",
  real: [],
  nextKey: 0,
};

const sessionPlanVersion = "0.1";
const seedSlot = "__seed__";

export function policyLines(text: string): readonly string[] {
  return text
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "");
}

function planOf(domain: string, draft: SessionDraft): Record<string, unknown> {
  const gateway = draft.gatewaySocket.trim();
  const policies = policyLines(draft.policies);
  return {
    schema_version: sessionPlanVersion,
    domain,
    seed: seedSlot,
    host_socket: draft.hostSocket.trim(),
    ...(gateway === "" ? {} : { gateway_socket: gateway }),
    ...(policies.length === 0 ? {} : { policies }),
    ...(draft.real.length === 0
      ? {}
      : { real: draft.real.map(({ instance, board, device }) => ({ instance, board, device })) }),
  };
}

export function sessionPlanText(domain: string, draft: SessionDraft): string {
  const seed = draft.seed.trim();
  const token = /^(0|[1-9]\d*)$/.test(seed) ? seed : JSON.stringify(draft.seed);
  return JSON.stringify(planOf(domain, draft)).replace(JSON.stringify(seedSlot), token);
}

export function withRealRow(draft: SessionDraft, instance: string): SessionDraft {
  const row = { key: draft.nextKey, instance, board: "", device: "" };
  return { ...draft, real: [...draft.real, row], nextKey: draft.nextKey + 1 };
}

export function withoutRealRow(draft: SessionDraft, key: number): SessionDraft {
  return { ...draft, real: draft.real.filter((row) => row.key !== key) };
}

export function withRowField(draft: SessionDraft, key: number, field: RowField, value: string): SessionDraft {
  return { ...draft, real: draft.real.map((row) => (row.key === key ? { ...row, [field]: value } : row)) };
}

function onRow(issue: PlanIssue, index: number, field: RowField): boolean {
  const row = `/real/${String(index)}`;
  return issue.pointer === `${row}/${field}` || (issue.pointer === row && issue.member === field);
}

export function rowIssues(issues: readonly PlanIssue[], index: number, field: RowField): readonly PlanIssue[] {
  return issues.filter((issue) => onRow(issue, index, field));
}

function onMember(issue: PlanIssue, member: PlanMember): boolean {
  return issue.pointer === `/${member}` || (issue.pointer === "/real" && issue.member === member);
}

export function memberIssues(issues: readonly PlanIssue[], member: PlanMember): readonly PlanIssue[] {
  return issues.filter((issue) => onMember(issue, member));
}

export function freeInstance(instances: readonly string[], rows: readonly RealRow[]): string | null {
  return instances.find((id) => !rows.some((row) => row.instance === id)) ?? null;
}
