import type { Outcome, PolicyDefaultFailure, PolicyDefaultFile } from "../../../app/ports";
import { isRecord } from "../../../model/shared/json";
import { names } from "./decodeKit";
import { fileContent, written } from "./decoders";
import { policyDefaultFailure } from "./failures";
import type { Transport } from "./http";
import { putFile, request } from "./http";

const defaultPath = "settings/default.policies.json";

function policiesIn(content: string): readonly string[] | null {
  try {
    const parsed = JSON.parse(content) as unknown;
    const only = isRecord(parsed) && Object.keys(parsed).join() === "policies";
    return only ? names(parsed.policies) : null;
  } catch {
    return null;
  }
}

function defaultFile(value: unknown): PolicyDefaultFile | null {
  const file = fileContent(value);
  if (file === null) return null;
  const policies = policiesIn(file.content);
  return policies === null ? { state: "unreadable", etag: file.etag } : { state: "present", policies, etag: file.etag };
}

export function readPolicyDefault(transport: Transport): Promise<Outcome<PolicyDefaultFile, PolicyDefaultFailure>> {
  return request(
    transport,
    `/api/file?path=${encodeURIComponent(defaultPath)}`,
    defaultFile,
    policyDefaultFailure,
    undefined,
    (problem) => (problem.code === "not_found" ? { state: "absent" } : null),
  );
}

function savedEtag(value: unknown): { readonly etag: string } | null {
  const file = written(value);
  return file === null ? null : { etag: file.etag };
}

export function savePolicyDefault(
  transport: Transport,
  policies: readonly string[],
  etag: string,
): Promise<Outcome<{ readonly etag: string }, PolicyDefaultFailure>> {
  const content = `${JSON.stringify({ policies }, null, 2)}\n`;
  return putFile(transport, defaultPath, content, etag, savedEtag, policyDefaultFailure);
}
