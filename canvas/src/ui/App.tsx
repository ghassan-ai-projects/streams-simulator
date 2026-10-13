import type { ThemeStore } from "../app/themeStore";
import { EmptyWorkspace } from "./shell/EmptyWorkspace";
import { Shell } from "./shell/Shell";
import { Sidebar } from "./shell/Sidebar";
import { TopBar } from "./shell/TopBar";
import { useTheme } from "./useTheme";

interface AppProps {
  readonly themes: ThemeStore;
}

export function App({ themes }: AppProps) {
  const { preference } = useTheme(themes);
  return (
    <Shell
      topBar={
        <TopBar
          preference={preference}
          onCycleTheme={() => {
            themes.cycle();
          }}
        />
      }
      sidebar={<Sidebar />}
    >
      <EmptyWorkspace />
    </Shell>
  );
}
