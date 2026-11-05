export interface OperatorCall {
  readonly atNs: number;
  readonly entity: string;
  readonly effector: string;
  readonly accepted: boolean;
  readonly reason: string;
}

export class CallLog {
  private calls: OperatorCall[] = [];

  constructor(private readonly capacity: number) {}

  add(call: OperatorCall): void {
    this.calls.push(call);
    if (this.calls.length >= this.capacity * 2) this.calls = this.calls.slice(-this.capacity);
  }

  between(fromNs: number, toNs: number): readonly OperatorCall[] {
    return this.calls.filter((call) => call.atNs >= fromNs && call.atNs <= toNs);
  }
}
