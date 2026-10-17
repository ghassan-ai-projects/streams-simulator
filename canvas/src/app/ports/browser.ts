export interface PreferenceStore {
  read(key: string): string | null;
  write(key: string, value: string): void;
}

export interface SystemTheme {
  prefersDark(): boolean;
  subscribe(listener: () => void): () => void;
}

export interface Location {
  read(): string;
  write(hash: string, replace?: boolean): void;
  subscribe(listener: () => void): () => void;
}
