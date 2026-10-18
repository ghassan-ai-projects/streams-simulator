import {
  blankKind,
  devicePath,
  parseChoices,
  parseKind,
  serialiseKind,
  stemOf,
  twinDomainPath,
  twinPath,
} from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceFailure, DevicesGateway, Outcome } from "../../ports";
import { failure, success } from "../../ports";
import type { CaptureSession, DeviceDesk, DeviceRow } from "./api";
import { createCaptureSession } from "./captureStore";
import { refusalWords } from "./deviceFailures";

function exists(name: string): DeviceFailure {
  return {
    kind: "conflict",
    diagnostic: { displayCode: "device_exists", message: `A device named ${name} already exists.` },
  };
}

function plain(kind: KindFile, refused: DeviceFailure): DeviceFailure {
  const message = refusalWords(kind, refused);
  return { ...refused, diagnostic: { ...refused.diagnostic, message } };
}

class Desk implements DeviceDesk {
  private session: CaptureSession | null = null;

  constructor(private readonly gateway: DevicesGateway) {}

  async rows(entries: readonly KindEntry[], domains: readonly string[]): Promise<readonly DeviceRow[]> {
    const rows = await Promise.all(entries.map((entry) => this.row(entry, domains)));
    return rows.sort((a, b) => a.name.localeCompare(b.name));
  }

  async create(name: string): Promise<Outcome<string, DeviceFailure>> {
    const stem = name.trim();
    const blank = blankKind(stem);
    const content = serialiseKind(blank);
    const checked = await this.gateway.checkContent(content, false);
    if (!checked.ok) return failure(plain(blank, checked.failure));
    const written = await this.gateway.writeFile(devicePath(stem), content, "");
    if (written.ok) return success(written.value.path);
    return failure(written.failure.kind === "conflict" ? exists(stem) : written.failure);
  }

  capture(): CaptureSession {
    this.session ??= createCaptureSession(this.gateway);
    return this.session;
  }

  private async row(entry: KindEntry, domains: readonly string[]): Promise<DeviceRow> {
    const file = await this.gateway.readFile(entry.path);
    const kind = file.ok ? parseKind(file.value.content) : null;
    return {
      path: entry.path,
      name: entry.kind === "" ? stemOf(entry.path) : entry.kind,
      metrics: kind === null ? null : kind.metrics.length,
      commands: kind === null ? null : kind.commands.length,
      status: kind === null || entry.kind === "" ? "invalid" : await this.twinStatus(entry, domains),
    };
  }

  private async twinStatus(entry: KindEntry, domains: readonly string[]): Promise<DeviceRow["status"]> {
    const file = await this.gateway.readFile(twinPath(stemOf(entry.path)));
    const choices = file.ok ? parseChoices(file.value.content) : null;
    return choices !== null && domains.includes(twinDomainPath(choices.id)) ? "twin-ready" : "needs-twin";
  }
}

export function createDeviceDesk(gateway: DevicesGateway): DeviceDesk {
  return new Desk(gateway);
}
