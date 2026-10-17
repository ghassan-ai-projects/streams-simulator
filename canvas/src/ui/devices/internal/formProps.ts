import type { DeviceSession } from "../../../app/devices";
import type { Command, FieldIssue, Metric } from "../../../model/devices";

export interface FormProps {
  readonly devices: DeviceSession;
  readonly issues: readonly FieldIssue[];
}

export type MetricEdit = Partial<Omit<Metric, "key">>;

export type CommandEdit = Partial<Omit<Command, "key">>;
