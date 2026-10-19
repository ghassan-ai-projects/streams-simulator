export interface NodeInterrupt {
  readonly mode: "hold" | "approve";
  readonly latched: boolean;
  readonly words: string;
}
