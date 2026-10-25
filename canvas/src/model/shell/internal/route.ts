export type Route =
  | { readonly screen: "workspace" }
  | { readonly screen: "plant"; readonly domain: string; readonly hall: string | null; readonly launch?: true }
  | { readonly screen: "review"; readonly worldId: string; readonly model?: true }
  | { readonly screen: "edit"; readonly composition: string }
  | { readonly screen: "device"; readonly path: string; readonly step?: "behaviour" }
  | { readonly screen: "capture" };

function plantRoute(params: URLSearchParams): Route | null {
  const domain = params.get("domain");
  if (!domain) return null;
  const hall = params.get("hall");
  return params.get("launch") === "1"
    ? { screen: "plant", domain, hall, launch: true }
    : { screen: "plant", domain, hall };
}

function fileRoute(path: string, params: URLSearchParams): Route | null {
  const world = params.get("world");
  const composition = params.get("composition");
  const device = params.get("device");
  if (path === "review" && world) {
    return params.get("model") === "1"
      ? { screen: "review", worldId: world, model: true }
      : { screen: "review", worldId: world };
  }
  if (path === "edit" && composition) return { screen: "edit", composition };
  if (path === "device" && device) {
    return params.get("step") === "behaviour"
      ? { screen: "device", path: device, step: "behaviour" }
      : { screen: "device", path: device };
  }
  return null;
}

export function parseRoute(hash: string): Route {
  const [path = "", query = ""] = hash.replace(/^#\/?/, "").split("?");
  const params = new URLSearchParams(query);
  if (path === "capture") return { screen: "capture" };
  const found = path === "plant" ? plantRoute(params) : fileRoute(path, params);
  return found ?? { screen: "workspace" };
}

function plantHash(route: Extract<Route, { screen: "plant" }>): string {
  const params = new URLSearchParams({ domain: route.domain });
  if (route.hall !== null) params.set("hall", route.hall);
  if (route.launch) params.set("launch", "1");
  return `#/plant?${params.toString()}`;
}

function reviewHash(route: Extract<Route, { screen: "review" }>): string {
  const params = new URLSearchParams({ world: route.worldId });
  if (route.model) params.set("model", "1");
  return `#/review?${params.toString()}`;
}

function deviceHash(route: Extract<Route, { screen: "device" }>): string {
  const params = new URLSearchParams({ device: route.path });
  if (route.step !== undefined) params.set("step", route.step);
  return `#/device?${params.toString()}`;
}

export function formatRoute(route: Route): string {
  switch (route.screen) {
    case "workspace":
      return "#/";
    case "plant":
      return plantHash(route);
    case "review":
      return reviewHash(route);
    case "edit":
      return `#/edit?${new URLSearchParams({ composition: route.composition }).toString()}`;
    case "device":
      return deviceHash(route);
    case "capture":
      return "#/capture";
  }
}
