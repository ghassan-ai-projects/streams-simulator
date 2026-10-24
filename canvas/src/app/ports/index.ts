export type {
  BoardCapture,
  CapturedKind,
  CaptureSource,
  DeviceFailure,
  DevicesGateway,
  DomainChecked,
  InstrumentQuery,
  KindChecked,
  LibraryInstrument,
  LintReport,
  TwinMade,
} from "./devices";
export type { EditorFailure, EditorGateway } from "./editor";
export type {
  ConnectFailure,
  ConnectGateway,
  ConnectStatus,
  FaultCategories,
  FaultCategory,
  ImmediateCall,
  InjectedFault,
  LaunchFailure,
  LaunchStep,
  JobRequest,
  JobStatus,
  LaunchGateway,
  PerturbationKind,
  PerturbationParamSpec,
  PlannedInstance,
  PolicyDefaultFailure,
  PolicyDefaultFile,
  SessionCheck,
} from "./launch";
export type { PlantFailure, PlantGateway } from "./plant";
export type { ControlFailure, DocumentFailure, ReviewFailure, ReviewGateway, RunFollower, TwinFailure } from "./review";
export type { SessionFailure, SessionGateway } from "./session";
export type { WorkspaceFailure, WorkspaceGateway } from "./workspace";
export type { Clock, Scheduler } from "./clock";
export type { Clipboard, FileSaver, Location, PreferenceStore, SystemTheme } from "./browser";
export { failure, success } from "./shared";
export type { ComposeSummary, Diagnostic, DomainDocument, FileContent, Outcome, PlantSource } from "./shared";

import type { DevicesGateway } from "./devices";
import type { EditorGateway } from "./editor";
import type { LaunchGateway } from "./launch";
import type { PlantGateway } from "./plant";
import type { ReviewGateway } from "./review";
import type { SessionGateway } from "./session";
import type { WorkspaceGateway } from "./workspace";

export interface BridgeAdapters {
  readonly session: SessionGateway;
  readonly workspace: WorkspaceGateway;
  readonly plant: PlantGateway;
  readonly editor: EditorGateway;
  readonly devices: DevicesGateway;
  readonly launch: LaunchGateway;
  readonly review: ReviewGateway;
}
