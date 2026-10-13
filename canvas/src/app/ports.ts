import type { FileContent, Outcome, Workspace } from "../model/bridge";

export interface PreferenceStore {
  read(key: string): string | null;
  write(key: string, value: string): void;
}

export interface SystemTheme {
  prefersDark(): boolean;
  subscribe(listener: () => void): () => void;
}

export interface Bridge {
  session(): Promise<Outcome<boolean>>;
  pair(code: string): Promise<Outcome<boolean>>;
  workspace(): Promise<Outcome<Workspace>>;
  readFile(path: string): Promise<Outcome<FileContent>>;
  action(name: string, args: Readonly<Record<string, unknown>>): Promise<Outcome<unknown>>;
  watchWorkspace(onChange: () => void): () => void;
  followRun(worldId: string, follower: RunFollower): () => void;
}

export interface RunFollower {
  record(data: string): void;
  closed(): void;
  failed(): void;
}

export interface Location {
  read(): string;
  write(hash: string): void;
  subscribe(listener: () => void): () => void;
}
