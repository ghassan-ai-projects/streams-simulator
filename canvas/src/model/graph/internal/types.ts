import type { Category, ChannelRange, ChannelState, Family, Instance, Pins } from "../../plant";
import type { Quality } from "../../shared/quality";

export interface ReadingValue {
  readonly channel: string;
  readonly value: number | string | boolean;
  readonly unit: string;
  readonly quality?: Quality;
}

export type RowCount = (pins: Pins, channels: readonly string[]) => number;

export interface ViewPresentation {
  readonly rows: RowCount;
  readonly note: (lines: number) => number;
}

export type Attention = "alarm" | "warn" | null;

export type NodeStatus = Attention | "fault";

export interface Shown {
  readonly channel: string;
  readonly value: string;
  readonly unit: string;
  readonly attention: Attention;
  readonly word: boolean;
  readonly outside: string;
  readonly quality: string;
}

export interface ChannelSpec {
  readonly states: readonly ChannelState[];
  readonly range: ChannelRange | null;
}

export interface ReadoutRow {
  readonly key: string;
  readonly label: string;
  readonly present: boolean;
  readonly value: string;
  readonly unit: string;
  readonly word: boolean;
  readonly attention: Attention;
  readonly outside: string;
  readonly observed: boolean;
  readonly quality: string;
  readonly absent: string;
}

export interface ReadoutBody {
  readonly rows: readonly ReadoutRow[];
  readonly hidden: number;
  readonly hiddenAttention: Attention;
}

export interface ViewNode {
  readonly id: string;
  readonly label: string;
  readonly detail: string;
  readonly group: string;
  readonly instance: Instance | null;
  readonly category: Category;
  readonly members: number;
  readonly pins: Pins;
  readonly channels: readonly string[];
  readonly noteLines: number;
}

export interface ViewEdge {
  readonly id: string;
  readonly from: string;
  readonly to: string;
  readonly fromHandle: string | null;
  readonly toHandle: string | null;
  readonly family: Family | null;
  readonly count: number;
}

export interface PlantView {
  readonly nodes: readonly ViewNode[];
  readonly edges: readonly ViewEdge[];
  readonly layout: Layout;
  readonly halls: readonly string[];
}

export interface Rect {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

export interface Viewport {
  readonly x: number;
  readonly y: number;
  readonly zoom: number;
}

export type CutSide = "none" | "left" | "right" | "both";

export interface PlanMark {
  readonly status: "fault" | null;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface LayoutNode {
  readonly id: string;
  readonly group: string;
  readonly width: number;
  readonly height: number;
}

export interface LayoutEdge {
  readonly from: string;
  readonly to: string;
}

export interface Layout {
  readonly nodes: ReadonlyMap<string, Rect>;
  readonly groups: ReadonlyMap<string, Rect>;
}

export interface FittedId {
  readonly text: string;
  readonly size: number;
}

export interface OutlineItem {
  readonly id: string;
  readonly kind: string;
  readonly flagged?: boolean;
}

export interface OutlineGroup {
  readonly area: string;
  readonly depth: number;
  readonly items: readonly OutlineItem[];
}
