import {
  choicesOfDocument,
  keyChoices,
  parseChoices,
  parseKind,
  reconcileChoices,
  serialiseChoices,
  stemOf,
  twinDomainPath,
  twinPath,
} from "../../../model/devices";
import type { KindFile, Member, TwinChoices } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceFailure, DevicesGateway, FileContent, Outcome } from "../../ports";
import { failure, success } from "../../ports";
import { unreadable } from "./deviceFailures";
import { loadCatalogue } from "./instrumentCatalogue";
import type { Catalogue, Held, TwinDraft } from "./twinDraft";
import type { Suggestion } from "./twinSuggest";
import { suggestFor } from "./twinSuggest";

export interface Loaded {
  readonly choicesPath: string;
  readonly members: readonly Member[];
  readonly known: readonly KindEntry[];
  readonly instruments: Catalogue;
  readonly suggested: Readonly<Record<string, unknown>> | null;
  readonly saved: TwinChoices;
  readonly written: string;
  readonly doc: TwinChoices;
  readonly touched: boolean;
  readonly onDisk: boolean;
  readonly etag: string;
  readonly domainEtag: string;
  readonly held: readonly Held[];
}

interface LoadFailure {
  readonly failure: DeviceFailure;
  readonly held: readonly Held[];
}

export async function readMember(gateway: DevicesGateway, path: string): Promise<Outcome<Member, DeviceFailure>> {
  const file = await gateway.readFile(path);
  if (!file.ok) return failure(file.failure);
  const kind = parseKind(file.value.content);
  return kind === null ? failure(unreadable) : success({ path, kind });
}

async function readOthers(
  gateway: DevicesGateway,
  known: readonly KindEntry[],
  primary: Member,
  stored: TwinChoices | null,
): Promise<readonly Member[]> {
  const names = (stored?.kinds ?? []).map((choice) => choice.kind).filter((name) => name !== primary.kind.kind);
  const entries = names.flatMap((name) => known.filter((entry) => entry.kind === name && entry.path !== primary.path));
  const read = await Promise.all(entries.map((entry) => readMember(gateway, entry.path)));
  return read.flatMap((member) => (member.ok ? [member.value] : []));
}

function settledChoices(
  base: TwinChoices,
  members: readonly Member[],
  suggested: Readonly<Record<string, unknown>> | null,
): TwinChoices {
  const fresh = suggested === null ? null : keyChoices(choicesOfDocument(suggested), members);
  return reconcileChoices(keyChoices(base, members), members, fresh);
}

async function domainEtagOf(gateway: DevicesGateway, domains: readonly string[], id: string): Promise<string> {
  const path = twinDomainPath(id);
  if (!domains.includes(path)) return "";
  const file = await gateway.readFile(path);
  return file.ok ? file.value.etag : "";
}

interface StoredFile {
  readonly stored: TwinChoices | null;
  readonly etag: string;
}

function storedOf(file: Outcome<FileContent, DeviceFailure>): Outcome<StoredFile, DeviceFailure> {
  if (file.ok) return success({ stored: parseChoices(file.value.content), etag: file.value.etag });
  return file.failure.kind === "missing" ? success({ stored: null, etag: "" }) : failure(file.failure);
}

function baseOf(stored: TwinChoices | null, suggestion: Suggestion): Outcome<TwinChoices, LoadFailure> {
  if (stored !== null) return success(stored);
  const { outcome, held } = suggestion;
  return outcome.ok ? success(choicesOfDocument(outcome.value)) : failure({ failure: outcome.failure, held });
}

export async function loadTwin(
  gateway: DevicesGateway,
  path: string,
  primary: KindFile,
): Promise<Outcome<Loaded, LoadFailure>> {
  const choicesPath = twinPath(stemOf(path));
  const [file, listing] = await Promise.all([gateway.readFile(choicesPath), gateway.listing()]);
  const read = storedOf(file);
  if (!read.ok) return failure({ failure: read.failure, held: [] });
  const { stored, etag } = read.value;
  const known = listing.ok ? listing.value.kinds : [];
  const lead = { path, kind: primary };
  const members = [lead, ...(await readOthers(gateway, known, lead, stored))];
  const suggestion = await suggestFor(gateway, members);
  const base = baseOf(stored, suggestion);
  if (!base.ok) return failure(base.failure);
  const suggested = suggestion.outcome.ok ? suggestion.outcome.value : null;
  const doc = settledChoices(base.value, members, suggested);
  const domains = listing.ok ? listing.value.domains : [];
  const domainEtag = await domainEtagOf(gateway, domains, doc.id);
  const saved = stored === null ? doc : keyChoices(stored, members);
  return success({
    choicesPath,
    members,
    known,
    instruments: await loadCatalogue(gateway, members, {}),
    suggested,
    saved,
    written: serialiseChoices(saved, members),
    doc,
    touched: stored !== null,
    onDisk: domainEtag !== "",
    etag,
    domainEtag,
    held: suggestion.held,
  });
}

export function reconciled(draft: TwinDraft, members: readonly Member[]): TwinChoices {
  const fresh = draft.suggested === null ? null : keyChoices(choicesOfDocument(draft.suggested), members);
  return reconcileChoices(draft.doc, members, fresh);
}
