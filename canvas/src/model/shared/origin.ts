export type OriginKind = "real" | "simulated" | "unknown";

export const originLabels: Readonly<Record<OriginKind, string>> = {
  real: "Real",
  simulated: "Simulated",
  unknown: "Origin unknown",
};
