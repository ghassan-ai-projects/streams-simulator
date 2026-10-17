import type { Route } from "../model/route";
import { formatRoute, parseRoute } from "../model/route";
import { Observable } from "./shared";
import type { Location } from "./ports";

export interface LeaveAttempt {
  readonly route: Route;
  readonly replace: boolean;
}

class LeaveSlot extends Observable<LeaveAttempt | null> {
  hold(attempt: LeaveAttempt | null): void {
    this.set(attempt);
  }
}

export class NavigationStore extends Observable<Route> {
  readonly leaving = new LeaveSlot(null);
  private unsaved: { readonly id: symbol; readonly check: () => boolean } | null = null;

  constructor(private readonly location: Location) {
    super(parseRoute(location.read()));
    location.subscribe(() => {
      this.arrive(parseRoute(location.read()));
    });
  }

  guardUnsaved(check: () => boolean): () => void {
    const guard = { id: Symbol("unsaved-guard"), check };
    this.unsaved = guard;
    return () => {
      if (this.unsaved?.id === guard.id) this.unsaved = null;
    };
  }

  go(route: Route): void {
    this.request({ route, replace: false });
  }

  replace(route: Route): void {
    this.request({ route, replace: true });
  }

  leave(): void {
    const attempt = this.leaving.snapshot();
    this.leaving.hold(null);
    if (attempt !== null) this.commit(attempt);
  }

  stay(): void {
    this.leaving.hold(null);
  }

  private request(attempt: LeaveAttempt): void {
    if (this.blocked(attempt.route)) this.leaving.hold(attempt);
    else this.commit(attempt);
  }

  private arrive(route: Route): void {
    if (!this.blocked(route)) {
      this.set(route);
      return;
    }
    this.location.write(formatRoute(this.snapshot()), true);
    this.leaving.hold({ route, replace: false });
  }

  private commit({ route, replace }: LeaveAttempt): void {
    this.location.write(formatRoute(route), replace);
    this.set(route);
  }

  private blocked(route: Route): boolean {
    const current = this.snapshot();
    return current.screen === "edit" && formatRoute(current) !== formatRoute(route) && (this.unsaved?.check() ?? false);
  }
}
