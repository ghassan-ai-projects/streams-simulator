# Canvas dependencies

Exact versions, lockfile committed, installed with `npm ci --ignore-scripts`; `npm audit --audit-level=high` clean. Each
version was published at least two weeks before it was pinned.

## Runtime

| Package              | Why                                                             |
| -------------------- | --------------------------------------------------------------- |
| `react`, `react-dom` | UI framework (decision record: React with Vite)                 |
| `lucide-react`       | The one icon set: consistent stroke icons, tree-shaken per icon |

## Build and test

| Package                                               | Why                                                                                                                              |
| ----------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `typescript` (6.0)                                    | Strict types; 6.0 because `typescript-eslint` supports TypeScript below 6.1                                                      |
| `vite`, `@vitejs/plugin-react`                        | Development server and production build                                                                                          |
| `vitest`, `@vitest/coverage-v8`, `jsdom`              | Unit and component tests, coverage thresholds per layer                                                                          |
| `@testing-library/react`, `@testing-library/jest-dom` | Component tests through the accessibility tree                                                                                   |
| `axe-core`, `@axe-core/playwright`                    | WCAG checks in component and end-to-end tests                                                                                    |
| `@playwright/test`, `playwright-core`                 | End-to-end tests in a real browser; `playwright-core` pinned to the runner's version so `@axe-core/playwright` uses the same one |
| `eslint`, `@eslint/js`, `typescript-eslint`           | Type-aware strict lint, zero warnings                                                                                            |
| `eslint-plugin-react-hooks`                           | Hook rules                                                                                                                       |
| `eslint-plugin-sonarjs`                               | Cognitive complexity limit                                                                                                       |
| `@eslint-community/eslint-plugin-eslint-comments`     | Every disable comment names its rule and gives a reason                                                                          |
| `@vitest/eslint-plugin`, `eslint-plugin-playwright`   | No focused, disabled or sleeping tests                                                                                           |
| `prettier`                                            | The one formatter                                                                                                                |
| `dependency-cruiser`                                  | Layer rules (`model` ← `app` ← `adapters`; `ui` never imports `adapters`)                                                        |
| `jscpd`                                               | No duplicated block of 75 tokens or more                                                                                         |
| `knip`                                                | No unused file, export or dependency                                                                                             |
| `@types/node`, `@types/react`, `@types/react-dom`     | Types                                                                                                                            |

`eslint-plugin-jsx-a11y` is not used: it does not support ESLint 10 yet; accessibility is enforced by axe in every
component and end-to-end test.
