import type { Badge } from "../../../model/review";
import { badgeOf } from "../../../model/review";
import type { ReviewState } from "./reviewState";

export class BadgeBook {
  private readonly badges = new Map<string, Badge>();

  of(worldId: string): Badge | null {
    return this.badges.get(worldId) ?? null;
  }

  record({ worldId, report, verified }: ReviewState): void {
    if (worldId !== null && report.phase === "ready" && verified.phase !== "loading") {
      this.badges.set(worldId, badgeOf(report.value, verified));
    }
  }
}
