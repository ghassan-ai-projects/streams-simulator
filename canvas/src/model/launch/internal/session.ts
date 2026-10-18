export interface RealRow {
  readonly instance: string;
  readonly board: string;
  readonly device: string;
}

export interface SessionDraft {
  readonly seed: string;
  readonly hostSocket: string;
  readonly gatewaySocket: string;
  readonly real: readonly RealRow[];
}

export type RowField = keyof RealRow;

export type PlanMember = "seed" | "host_socket" | "gateway_socket";

export interface PlanIssue {
  readonly pointer: string;
  readonly message: string;
  readonly keyword?: string;
  readonly member?: string;
}

export type InstanceRole = "real" | "simulated";

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
}

export interface BindingNote {
  readonly matched: boolean;
  readonly refusal: string;
  readonly reason: string;
  readonly issues: readonly PlanIssue[];
}

export const firstSessionDraft: SessionDraft = { seed: "1", hostSocket: "", gatewaySocket: "", real: [] };

const emptyRow: RealRow = { instance: "", board: "", device: "" };

function seedValue(seed: string): number | string {
  return /^\d{1,15}$/.test(seed.trim()) ? Number(seed.trim()) : seed;
}

export function sessionPlanText(domain: string, draft: SessionDraft): string {
  const gateway = draft.gatewaySocket.trim();
  const plan = {
    schema_version: "0.1",
    domain,
    seed: seedValue(draft.seed),
    host_socket: draft.hostSocket.trim(),
    ...(gateway === "" ? {} : { gateway_socket: gateway }),
    ...(draft.real.length === 0 ? {} : { real: draft.real.map((row) => ({ ...row })) }),
  };
  return JSON.stringify(plan);
}

export function withRealRow(draft: SessionDraft, instance: string): SessionDraft {
  return { ...draft, real: [...draft.real, { ...emptyRow, instance }] };
}

export function withoutRealRow(draft: SessionDraft, index: number): SessionDraft {
  return { ...draft, real: draft.real.filter((_, at) => at !== index) };
}

export function withRowField(draft: SessionDraft, index: number, field: RowField, value: string): SessionDraft {
  return { ...draft, real: draft.real.map((row, at) => (at === index ? { ...row, [field]: value } : row)) };
}

export function rowIssue(issues: readonly PlanIssue[], index: number, field: RowField): PlanIssue | undefined {
  return issues.find((issue) => issue.pointer === `/real/${String(index)}/${field}`);
}

export function memberIssue(issues: readonly PlanIssue[], member: PlanMember): PlanIssue | undefined {
  return issues.find(
    (issue) => issue.pointer === `/${member}` || (issue.member === member && !issue.pointer.includes("/real/")),
  );
}

const placed = /^\/(seed|host_socket|gateway_socket|real\/\d+\/(instance|board|device))$/;

export function unplacedIssues(issues: readonly PlanIssue[]): readonly PlanIssue[] {
  return issues.filter((issue) => !placed.test(issue.pointer) && issue.member !== "gateway_socket");
}
