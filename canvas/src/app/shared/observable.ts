export interface Store<S> {
  readonly snapshot: () => S;
  readonly subscribe: (listener: () => void) => () => void;
}

export class Observable<S> {
  private readonly listeners = new Set<() => void>();

  constructor(private state: S) {}

  readonly snapshot = (): S => this.state;

  readonly subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  protected set(next: S): void {
    if (Object.is(next, this.state)) return;
    this.state = next;
    this.listeners.forEach((listener) => {
      listener();
    });
  }
}
