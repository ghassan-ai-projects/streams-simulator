export function wiresUnderNodes(): string[] {
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
  const found: string[] = [];
  document.querySelectorAll<SVGPathElement>(".react-flow__edge .react-flow__edge-path").forEach((path) => {
    const matrix = path.getScreenCTM();
    if (matrix === null) return;
    const length = path.getTotalLength();
    for (let at = 16; at < length - 16; at += 1) {
      const point = path.getPointAtLength(at);
      const hit = under(
        point.x * matrix.a + point.y * matrix.c + matrix.e,
        point.x * matrix.b + point.y * matrix.d + matrix.f,
      );
      if (hit !== undefined) {
        found.push(`${path.closest(".react-flow__edge")?.getAttribute("data-id") ?? "wire"} under ${hit.id}`);
        break;
      }
    }
  });
  return found;
}

export function wireEndsTooClose(): string[] {
  const ends: { readonly id: string; readonly x: number; readonly y: number }[] = [];
  document.querySelectorAll<SVGPathElement>(".react-flow__edge .react-flow__edge-path").forEach((path) => {
    const id = path.closest(".react-flow__edge")?.getAttribute("data-id") ?? "wire";
    const start = path.getPointAtLength(0);
    const finish = path.getPointAtLength(path.getTotalLength());
    ends.push({ id, x: start.x, y: start.y }, { id, x: finish.x, y: finish.y });
  });
  return ends.flatMap((a, index) =>
    ends
      .slice(index + 1)
      .filter((b) => a.id !== b.id && Math.hypot(a.x - b.x, a.y - b.y) < 6)
      .map((b) => `${a.id} meets ${b.id}`),
  );
}
