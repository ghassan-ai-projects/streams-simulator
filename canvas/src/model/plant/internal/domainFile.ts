import { isRecord, text } from "../../shared/json";

export interface DomainFile {
  readonly domain: Readonly<Record<string, unknown>>;
  readonly id: string;
}

export function parseDomainFile(content: string): DomainFile {
  try {
    const value: unknown = JSON.parse(content);
    return isRecord(value) ? { domain: value, id: text(value.id) } : { domain: {}, id: "" };
  } catch {
    return { domain: {}, id: "" };
  }
}
