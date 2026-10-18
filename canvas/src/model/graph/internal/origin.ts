export interface InstanceOrigin {
  readonly instance: string;
  readonly kind: "real" | "simulated";
  readonly board: string;
  readonly device: string;
  readonly bootId: string;
}

export const originLabels: Readonly<Record<InstanceOrigin["kind"], string>> = { real: "Real", simulated: "Simulated" };

export function originWords(origin: InstanceOrigin): string {
  const board = `${originLabels[origin.kind]} · board ${origin.board}`;
  return origin.kind === "real" && origin.bootId !== "" ? `${board} · boot ${origin.bootId}` : board;
}
