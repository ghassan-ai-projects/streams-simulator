import type { Route } from "../model/route";
import { formatRoute, parseRoute } from "../model/route";
import { Observable } from "./observable";
import type { Location } from "./ports";

export class NavigationStore extends Observable<Route> {
  constructor(private readonly location: Location) {
    super(parseRoute(location.read()));
    location.subscribe(() => {
      this.set(parseRoute(location.read()));
    });
  }

  go(route: Route): void {
    this.location.write(formatRoute(route));
    this.set(route);
  }

  replace(route: Route): void {
    this.location.write(formatRoute(route), true);
    this.set(route);
  }
}
