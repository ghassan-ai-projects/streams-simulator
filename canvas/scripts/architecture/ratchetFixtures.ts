import { shallowProblems } from "./ratchet.ts";
import { compare, register } from "./ratchetShape.ts";
import { runRenameFixtures } from "./ratchetRenameFixtures.ts";

const phase = "P1";
const oldNodes = "src/ui/plant/nodes.module.css";
const newNodes = "src/ui/graph/internal/nodes.module.css";
const oldControls = "src/ui/plant/GraphControls.module.css";
const vendor = "src/ui/graph/internal/vendor.module.css";

export function runRatchetFixtures(): string[] {
  return [
    ...validMoveFixture(),
    ...unrelatedMoveFixture(),
    ...runRenameFixtures(),
    ...vendorCapFixture(),
    ...vendorTighterFixture(),
    ...vendorSourceFixture(),
    ...crossLayerFixture(),
    ...shallowFixture(),
  ];
}

function shallowFixture(): string[] {
  const failing = [shallowProblems("true"), shallowProblems(null)];
  const passing = shallowProblems("false");
  if (failing.every((problems) => problems.length === 1) && passing.length === 0) return [];
  return ["ratchet a shallow or unknown clone must fail loudly and a full clone must pass"];
}

function validMoveFixture(): string[] {
  const base = register({
    caps: { [oldNodes]: { cap: 695, phase } },
    globalCaps: {
      [oldNodes]: { global: 70, important: 30, phase },
      [oldControls]: { global: 0, important: 2, phase },
    },
    debt: [
      { rule: "boolean-flag-param", file: "src/ui/plant/flow.ts", count: 1, phase },
      {
        rule: "css-feature-ownership",
        file: "src/ui/shared.module.css",
        target: "src/ui/plant/Outline.tsx",
        count: 1,
        phase,
      },
    ],
  });
  const moved = register({
    extracted: true,
    caps: { [newNodes]: { cap: 695, phase } },
    globalCaps: {
      [newNodes]: { global: 0, important: 30, phase },
      [vendor]: { global: 6, important: 2, phase },
    },
    vendorFiles: [vendor],
    debt: [
      { rule: "boolean-flag-param", file: "src/ui/graph/internal/flow.ts", count: 1, phase },
      {
        rule: "css-feature-ownership",
        file: "src/ui/shared.module.css",
        target: "src/ui/graph/internal/Outline.tsx",
        count: 1,
        phase,
      },
    ],
  });
  return compare("same-basename extraction and approved vendor move", base, moved, []);
}

function unrelatedMoveFixture(): string[] {
  const base = register({
    caps: { [oldNodes]: { cap: 695, phase } },
    globalCaps: { [oldNodes]: { global: 70, important: 30, phase } },
    debt: [{ rule: "boolean-flag-param", file: "src/ui/plant/flow.ts", count: 1, phase }],
  });
  const unrelated = register({
    extracted: true,
    caps: { "src/ui/graph/internal/renamed.module.css": { cap: 400, phase } },
    globalCaps: { "src/ui/graph/internal/other.module.css": { global: 0, important: 2, phase } },
    debt: [{ rule: "boolean-flag-param", file: "src/ui/graph/internal/other.ts", count: 1, phase }],
  });
  return compare("unrelated paths do not inherit moved allowances", base, unrelated, [
    "new :global/!important allowance src/ui/graph/internal/other.module.css",
    "new css cap src/ui/graph/internal/renamed.module.css is above the default 300",
    "new debt boolean-flag-param src/ui/graph/internal/other.ts",
  ]);
}

function vendorCapFixture(): string[] {
  const base = register({
    globalCaps: { [oldControls]: { global: 0, important: 2, phase } },
  });
  const badVendor = register({
    extracted: true,
    globalCaps: { [vendor]: { global: 7, important: 3, phase } },
    vendorFiles: [vendor],
  });
  return compare("vendor registration stays fixed to its approved allowance", base, badVendor, [
    "new :global/!important allowance src/ui/graph/internal/vendor.module.css",
    "vendor stylesheet allow-list gained src/ui/graph/internal/vendor.module.css",
  ]);
}

function vendorTighterFixture(): string[] {
  const base = register({ globalCaps: { [oldControls]: { global: 0, important: 2, phase } } });
  const tighter = register({
    extracted: true,
    globalCaps: { [vendor]: { global: 6, important: 0, phase } },
    vendorFiles: [vendor],
  });
  return compare("vendor registration may be tighter than its approved allowance", base, tighter, []);
}

function vendorSourceFixture(): string[] {
  const noSourceVendor = register({
    extracted: true,
    globalCaps: {
      [newNodes]: { global: 0, important: 30, phase },
      [vendor]: { global: 6, important: 2, phase },
    },
    vendorFiles: [vendor],
  });
  return compare(
    "vendor registration requires the moved source allowance",
    register({
      caps: { [oldNodes]: { cap: 695, phase } },
      globalCaps: { [oldNodes]: { global: 70, important: 30, phase } },
    }),
    noSourceVendor,
    [
      "new :global/!important allowance src/ui/graph/internal/vendor.module.css",
      "vendor stylesheet allow-list gained src/ui/graph/internal/vendor.module.css",
    ],
  );
}

function crossLayerFixture(): string[] {
  const base = register({
    caps: { [oldNodes]: { cap: 695, phase } },
    globalCaps: { [oldNodes]: { global: 0, important: 30, phase } },
  });
  const crossLayer = register({
    extracted: true,
    graphLayer: "model",
    caps: { [newNodes]: { cap: 695, phase } },
    globalCaps: { [newNodes]: { global: 0, important: 30, phase } },
  });
  return compare("cross-layer paths do not inherit moved allowances", base, crossLayer, [
    "new css cap src/ui/graph/internal/nodes.module.css is above the default 300",
    "new :global/!important allowance src/ui/graph/internal/nodes.module.css",
  ]);
}
