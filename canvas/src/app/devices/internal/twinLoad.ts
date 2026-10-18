import { parseChoices, parseKind, reconcileChoices, startingChoices, stemOf, twinPath } from "../../../model/devices";
import type { Instrument, TwinChoices } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceFailure, DevicesGateway, Outcome } from "../../ports";
import { failure, success } from "../../ports";
import { unreadable } from "./deviceFailures";
import { loadInstruments } from "./instrumentCatalogue";
import type { Member } from "./twinDraft";

export interface Loaded {
  readonly choicesPath: string;
  readonly members: readonly Member[];
  readonly known: readonly KindEntry[];
  readonly instruments: readonly Instrument[] | null;
  readonly saved: TwinChoices;
  readonly doc: TwinChoices;
  readonly touched: boolean;
  readonly etag: string;
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

export async function loadTwin(gateway: DevicesGateway, path: string): Promise<Outcome<Loaded, DeviceFailure>> {
  const choicesPath = twinPath(stemOf(path));
  const [primary, file, listing, catalogue] = await Promise.all([
    readMember(gateway, path),
    gateway.readFile(choicesPath),
    gateway.kinds(),
    loadInstruments(gateway),
  ]);
  if (!primary.ok) return primary;
  if (!file.ok && file.failure.kind !== "missing") return failure(file.failure);
  const known = listing.ok ? listing.value : [];
  const stored = file.ok ? parseChoices(file.value.content) : null;
  const members = [primary.value, ...(await readOthers(gateway, known, primary.value, stored))];
  const instruments = catalogue ?? [];
  const kinds = members.map((member) => member.kind);
  const doc = reconcileChoices(stored ?? startingChoices(kinds, instruments), kinds, instruments);
  const etag = file.ok ? file.value.etag : "";
  return success({
    choicesPath,
    members,
    known,
    instruments: catalogue,
    saved: stored ?? doc,
    doc,
    touched: stored !== null,
    etag,
  });
}
