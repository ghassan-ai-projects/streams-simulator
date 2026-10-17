import type { BridgeAdapters } from "../../../app/ports";
import { fileContent, written } from "./decoders";
import { deviceFailure } from "./failures";
import type { Transport } from "./http";
import { action, getFile, putFile } from "./http";
import { kindChecked } from "./kindIssues";

export function devicesGateway(transport: Transport): BridgeAdapters["devices"] {
  return {
    readFile: (path) => getFile(transport, path, fileContent, deviceFailure("read")),
    writeFile: (path, content, etag) => putFile(transport, path, content, etag, written, deviceFailure("write")),
    check: (path) => action(transport, "kind.validate", { path }, kindChecked, deviceFailure("check")),
  };
}
