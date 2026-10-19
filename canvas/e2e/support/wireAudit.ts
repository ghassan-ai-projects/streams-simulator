import type { Audit } from "./audit";

export function wiresUnderNodes(atLeast = 1): Audit {
  const graph = document.querySelector("[aria-roledescription='plant graph']");
  const zoom = Number(graph === null ? 1 : getComputedStyle(graph).getPropertyValue("--zoom"));
  const margin = 4 * zoom;
  const nodes = [...document.querySelectorAll<HTMLElement>(".react-flow__node-instance, .react-flow__node-hall")].map(
    (node) => ({ id: node.dataset.id ?? "", box: node.getBoundingClientRect() }),
  );
  const under = (x: number, y: number) =>
    nodes.find(
      ({ box }) => x > box.left - margin && x < box.right + margin && y > box.top - margin && y < box.bottom + margin,
    );
  const clearance = (x: number, y: number) =>
    Math.min(
      ...nodes.map(({ box }) =>
        Math.hypot(
          Math.max(box.left - margin - x, 0, x - box.right - margin),
          Math.max(box.top - margin - y, 0, y - box.bottom - margin),
        ),
      ),
    );
  const found: string[] = [];
  const paths = [...document.querySelectorAll<SVGPathElement>(".react-flow__edge .react-flow__edge-path")];
  paths.forEach((path) => {
    const matrix = path.getScreenCTM();
    if (matrix === null) return;
    const scale = Math.hypot(matrix.a, matrix.b, matrix.c, matrix.d);
    const length = path.getTotalLength();
    for (let at = 16; at < length - 16;) {
      const point = path.getPointAtLength(at);
      const x = point.x * matrix.a + point.y * matrix.c + matrix.e;
      const y = point.x * matrix.b + point.y * matrix.d + matrix.f;
      const hit = under(x, y);
      if (hit !== undefined) {
        found.push(`${path.closest(".react-flow__edge")?.getAttribute("data-id") ?? "wire"} under ${hit.id}`);
        break;
      }
      at += Math.max(1, Math.floor(clearance(x, y) / scale));
    }
  });
  return { considered: paths.length, problems: found, met: paths.length >= atLeast };
}

export function wireEndsTooClose(atLeast = 1): Audit {
  const ends: { readonly id: string; readonly port: string; readonly start: boolean; x: number; y: number }[] = [];
  document.querySelectorAll<SVGPathElement>(".react-flow__edge .react-flow__edge-path").forEach((path) => {
    const matrix = path.getScreenCTM();
    if (matrix === null) return;
    const id = path.closest(".react-flow__edge")?.getAttribute("data-id") ?? "wire";
    const [source = "", target = ""] = id.split(">");
    const at = (point: DOMPoint) => ({
      x: point.x * matrix.a + point.y * matrix.c + matrix.e,
      y: point.x * matrix.b + point.y * matrix.d + matrix.f,
    });
    ends.push(
      { id, port: source, start: true, ...at(path.getPointAtLength(0)) },
      { id, port: target, start: false, ...at(path.getPointAtLength(path.getTotalLength())) },
    );
  });
  const problems = ends.flatMap((a, index) =>
    ends
      .slice(index + 1)
      .filter((b) => a.id !== b.id && !(a.start && b.start && a.port === b.port))
      .filter((b) => Math.hypot(a.x - b.x, a.y - b.y) < 6 - 1e-6)
      .map((b) => `${a.id} meets ${b.id}`),
  );
  return { considered: ends.length, problems: problems, met: ends.length >= atLeast };
}

export function wiresNearTitles(atLeast = 1): Audit {
  const texts = [...document.querySelectorAll(".react-flow__node-frameTitle")].flatMap((title) => {
    const text = document.createTreeWalker(title, NodeFilter.SHOW_TEXT).nextNode();
    if (text === null) return [];
    const range = document.createRange();
    range.selectNodeContents(text);
    return [{ title: text.textContent ?? "", box: range.getBoundingClientRect() }];
  });
  const found = new Set<string>();
  const paths = [...document.querySelectorAll<SVGPathElement>(".react-flow__edge .react-flow__edge-path")];
  paths.forEach((path) => {
    const matrix = path.getScreenCTM();
    if (matrix === null) return;
    const id = path.closest(".react-flow__edge")?.getAttribute("data-id") ?? "wire";
    const scale = Math.hypot(matrix.a, matrix.b, matrix.c, matrix.d);
    for (let at = 0; at <= path.getTotalLength();) {
      const point = path.getPointAtLength(at);
      const x = point.x * matrix.a + point.y * matrix.c + matrix.e;
      const y = point.x * matrix.b + point.y * matrix.d + matrix.f;
      let nearest = Infinity;
      for (const { title, box } of texts) {
        const dx = Math.max(box.left - x, 0, x - box.right);
        const dy = Math.max(box.top - y, 0, y - box.bottom);
        nearest = Math.min(nearest, Math.hypot(dx, dy));
        if (Math.hypot(dx, dy) < 8) found.add(`${id} near ${title}`);
      }
      at += 3 * Math.max(1, Math.floor((nearest - 8) / (3 * scale)));
    }
  });
  const problems = [...found];
  return { considered: paths.length, problems: problems, met: paths.length >= atLeast };
}
