export interface AlarmActivation {
  readonly atNs: number;
  readonly entity: string;
  readonly state: string;
  readonly priority: string;
}
