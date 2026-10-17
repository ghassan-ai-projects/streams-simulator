/** Layers: model (pure types and rules) <- app (state, use cases, ports) <- adapters (bridge, files) ; ui (components) uses app and model only. */
module.exports = {
  forbidden: [
    { name: "model-is-pure", severity: "error", from: { path: "^src/model" }, to: { path: "^src/(app|adapters|ui)" } },
    {
      name: "model-imports-nothing",
      severity: "error",
      from: { path: "^src/model" },
      to: { dependencyTypes: ["npm"] },
    },
    {
      name: "app-below-adapters-and-ui",
      severity: "error",
      from: { path: "^src/app" },
      to: { path: "^src/(adapters|ui)" },
    },
    { name: "adapters-below-ui", severity: "error", from: { path: "^src/adapters" }, to: { path: "^src/ui" } },
    { name: "ui-never-touches-transport", severity: "error", from: { path: "^src/ui" }, to: { path: "^src/adapters" } },
    {
      name: "no-test-support-in-production",
      severity: "error",
      from: { pathNot: ["^src/test/"] },
      to: { path: "^src/test/" },
    },
    { name: "no-circular", severity: "error", from: {}, to: { circular: true } },
    {
      name: "no-orphans",
      severity: "error",
      from: { orphan: true, pathNot: ["\\.d\\.ts$", "\\.test\\.tsx?$", "^src/test/"] },
      to: {},
    },
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    tsConfig: { fileName: "tsconfig.app.json" },
    tsPreCompilationDeps: true,
    exclude: { path: "\\.test\\.tsx?$" },
  },
};
