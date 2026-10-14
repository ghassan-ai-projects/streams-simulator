export type Route =
  | { readonly screen: "home" }
  | { readonly screen: "plant"; readonly domain: string; readonly hall: string | null }
  | { readonly screen: "run"; readonly worldId: string }
  | { readonly screen: "edit"; readonly composition: string };

export function parseRoute(hash: string): Route {
  const [path = "", query = ""] = hash.replace(/^#\/?/, "").split("?");
  const params = new URLSearchParams(query);
  const domain = params.get("domain");
  const world = params.get("world");
  const composition = params.get("composition");
  if (path === "plant" && domain) {
    return { screen: "plant", domain, hall: params.get("hall") };
  }
  if (path === "run" && world) {
    return { screen: "run", worldId: world };
  }
  if (path === "edit" && composition) {
    return { screen: "edit", composition };
  }
  return { screen: "home" };
}

export function formatRoute(route: Route): string {
  switch (route.screen) {
    case "home":
      return "#/";
    case "plant": {
      const params = new URLSearchParams({ domain: route.domain });
      if (route.hall !== null) params.set("hall", route.hall);
      return `#/plant?${params.toString()}`;
    }
    case "run":
      return `#/run?${new URLSearchParams({ world: route.worldId }).toString()}`;
    case "edit":
      return `#/edit?${new URLSearchParams({ composition: route.composition }).toString()}`;
  }
}
