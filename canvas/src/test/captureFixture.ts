import { failure as failed, success } from "./outcome";
import type { FakeBridge } from "./fakeBridge";
import { cannedFailure } from "./fakeFailures";

export const pumpDraft = {
  kind: "pump",
  metrics: [
    {
      name: "flow",
      data_type: "number",
      unit: "m3/h",
      eu_range: { min: 0, max: 120 },
      resolution: 0.1,
      writable: false,
    },
    { name: "speed", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, writable: true },
  ],
  commands: [
    {
      name: "set_speed",
      risk_class: "medium",
      safe_stop: false,
      args_schema: {
        type: "object",
        required: ["pct"],
        properties: { pct: { type: "number", minimum: 0, maximum: 100 } },
      },
    },
    { name: "safe_stop", risk_class: "low", safe_stop: true, args_schema: { type: "object" } },
  ],
};

export function announced(draft: { readonly kind: string } & Record<string, unknown> = pumpDraft, extra: object = {}) {
  return { kind: draft.kind, kind_digest: "sha256:aaa", draft, ...extra };
}

export function captureResult(kinds: readonly object[] = [announced()], over: object = {}) {
  return { environment: "hardware", bd_seq: 2, kinds, ...over };
}

export function differs(at = "/metrics/0/unit") {
  const comparison = { status: "differs", file: "kinds/pump.kind.json", file_digest: "sha256:old", differs_at: at };
  return announced(pumpDraft, { comparison });
}

export function same() {
  return announced(pumpDraft, {
    comparison: { status: "same", file: "kinds/pump.kind.json", file_digest: "sha256:aaa" },
  });
}

export function serveCapture(bridge: FakeBridge, result: object = captureResult()): void {
  bridge.actions.set("kind.capture", success(result));
}

export function refuseCapture(
  bridge: FakeBridge,
  keyword = "kind_digest_mismatch",
  message = "the birth is refused",
): void {
  const issues = [{ pointer: "/kinds/0/kind_digest", message: "does not hold", keyword }];
  bridge.actions.set("kind.capture", failed({ ...cannedFailure("kind_failed", 422, message), issues }));
}
