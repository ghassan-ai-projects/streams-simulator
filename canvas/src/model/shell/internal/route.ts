export type Route =
  | { readonly screen: "workspace" }
  | { readonly screen: "plant"; readonly domain: string; readonly hall: string | null; readonly launch?: true }
  | { readonly screen: "review"; readonly worldId: string }
  | { readonly screen: "edit"; readonly composition: string }
  | { readonly screen: "device"; readonly path: string }
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
  if (path === "review" && world) return { screen: "review", worldId: world };
  if (path === "edit" && composition) return { screen: "edit", composition };
  if (path === "device" && device) return { screen: "device", path: device };
  return null;
}

export function parseRoute(hash: string): Route {
  const [path = "", query = ""] = hash.replace(/^#\/?/, "").split("?");
  const params = new URLSearchParams(query);
  if (path === "capture") return { screen: "capture" };
  const found = path === "plant" ? plantRoute(params) : fileRoute(path, params);
  return found ?? { screen: "workspace" };
}

export function formatRoute(route: Route): string {
  switch (route.screen) {
    case "workspace":
      return "#/";
    case "plant": {
      const params = new URLSearchParams({ domain: route.domain });
      if (route.hall !== null) params.set("hall", route.hall);
      if (route.launch) params.set("launch", "1");
      return `#/plant?${params.toString()}`;
    }
    case "review":
      return `#/review?${new URLSearchParams({ world: route.worldId }).toString()}`;
    case "edit":
      return `#/edit?${new URLSearchParams({ composition: route.composition }).toString()}`;
    case "device":
      return `#/device?${new URLSearchParams({ device: route.path }).toString()}`;
    case "capture":
      return "#/capture";
  }
}
