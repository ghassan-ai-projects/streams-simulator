import type { Outcome } from "../../../app/ports/index";
import { failure, success } from "../../../app/ports/shared";

export interface Transport {
  fetch(input: string, init?: RequestInit): Promise<Response>;
  events(url: string): EventSource;
}

export interface WireIssue {
  readonly pointer: string;
  readonly message: string;
}

export interface WireProblem {
  readonly status: number;
  readonly code: string;
  readonly message: string;
  readonly issues: readonly WireIssue[];
  readonly result: unknown;
}

export type RequestFailure =
  | { readonly kind: "transport" }
  | { readonly kind: "http"; readonly problem: WireProblem }
  | { readonly kind: "unreadable" };

export type Decoder<T> = (value: unknown) => T | null;
export type FailureMapper<F> = (problem: RequestFailure) => F;

export async function request<T, F>(
  transport: Transport,
  url: string,
  decoder: Decoder<T>,
  mapFailure: FailureMapper<F>,
  init?: RequestInit,
  recoverHttp?: (problem: WireProblem) => T | null,
): Promise<Outcome<T, F>> {
  let response: Response;
  try {
    response = await transport.fetch(url, { credentials: "same-origin", ...init });
  } catch {
    return failure(mapFailure({ kind: "transport" }));
  }
  const payload = await responseBody(response);
  if (!response.ok) {
    const problem = wireProblem(response.status, payload);
    const recovered = recoverHttp?.(problem);
    return recovered === null || recovered === undefined
      ? failure(mapFailure({ kind: "http", problem }))
      : success(recovered);
  }
  const decoded = decoder(payload);
  return decoded === null ? failure(mapFailure({ kind: "unreadable" })) : success(decoded);
}

export function action<T, F>(
  transport: Transport,
  name: string,
  args: Readonly<Record<string, unknown>>,
  decoder: Decoder<T>,
  mapFailure: FailureMapper<F>,
  recoverHttp?: (problem: WireProblem) => T | null,
): Promise<Outcome<T, F>> {
  return request(
    transport,
    `/api/actions/${encodeURIComponent(name)}`,
    actionResult(decoder),
    mapFailure,
    post({ args }),
    recoverHttp === undefined ? undefined : (problem) => recoverHttp(problem),
  );
}

export function getFile<T, F>(
  transport: Transport,
  path: string,
  decoder: Decoder<T>,
  mapFailure: FailureMapper<F>,
): Promise<Outcome<T, F>> {
  return request(transport, `/api/file?path=${encodeURIComponent(path)}`, decoder, mapFailure);
}

export function putFile<T, F>(
  transport: Transport,
  path: string,
  content: string,
  etag: string,
  decoder: Decoder<T>,
  mapFailure: FailureMapper<F>,
): Promise<Outcome<T, F>> {
  return request(transport, "/api/file", decoder, mapFailure, send("PUT", { path, content, etag }));
}

function actionResult<T>(decoder: Decoder<T>): Decoder<T> {
  return (value) => (isRecord(value) && "result" in value ? decoder(value.result) : null);
}

function post(payload: unknown): RequestInit {
  return send("POST", payload);
}

function send(method: "POST" | "PUT", payload: unknown): RequestInit {
  return { method, headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) };
}

async function responseBody(response: Response): Promise<unknown> {
  try {
    return (await response.json()) as unknown;
  } catch {
    return null;
  }
}

function wireProblem(status: number, value: unknown): WireProblem {
  const body = isRecord(value) && isRecord(value.error) ? value.error : {};
  return {
    status,
    code: string(body.code) || "bridge_failed",
    message: string(body.message) || `The bridge answered ${String(status)}.`,
    issues: issueList(body.issues),
    result: body.result ?? null,
  };
}

function issueList(value: unknown): readonly WireIssue[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry) || typeof entry.pointer !== "string" || typeof entry.message !== "string") return [];
    return [{ pointer: entry.pointer, message: entry.message }];
  });
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function string(value: unknown): string {
  return typeof value === "string" ? value : "";
}
