import { blankKind, devicePath, parseKind, serialiseKind, stemOf } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceFailure, DevicesGateway, Outcome } from "../../ports";
import { failure, success } from "../../ports";
import type { DeviceDesk, DeviceRow } from "./api";

function exists(name: string): DeviceFailure {
  return {
    kind: "conflict",
    diagnostic: { displayCode: "device_exists", message: `A device named ${name} already exists.` },
  };
}

class Desk implements DeviceDesk {
  constructor(private readonly gateway: DevicesGateway) {}

  async rows(entries: readonly KindEntry[]): Promise<readonly DeviceRow[]> {
    const rows = await Promise.all(entries.map((entry) => this.row(entry)));
    return rows.sort((a, b) => a.name.localeCompare(b.name));
  }

  async create(name: string): Promise<Outcome<string, DeviceFailure>> {
    const stem = name.trim();
    const written = await this.gateway.writeFile(devicePath(stem), serialiseKind(blankKind(stem)), "");
    if (written.ok) return success(written.value.path);
    return failure(written.failure.kind === "conflict" ? exists(stem) : written.failure);
  }

  private async row(entry: KindEntry): Promise<DeviceRow> {
    const file = await this.gateway.readFile(entry.path);
    const kind = file.ok ? parseKind(file.value.content) : null;
    return {
      path: entry.path,
      name: entry.kind === "" ? stemOf(entry.path) : entry.kind,
      metrics: kind === null ? null : kind.metrics.length,
      commands: kind === null ? null : kind.commands.length,
    };
  }
}

export function createDeviceDesk(gateway: DevicesGateway): DeviceDesk {
  return new Desk(gateway);
}
