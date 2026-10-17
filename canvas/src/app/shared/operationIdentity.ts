export interface Operation {
  live(): boolean;
}

export interface OperationIdentity {
  begin(): Operation;
  invalidate(): void;
  current(): Operation | null;
}

export function createOperationIdentity(): OperationIdentity {
  let active: Operation | null = null;
  return {
    begin() {
      const operation: Operation = { live: () => active === operation };
      active = operation;
      return operation;
    },
    invalidate() {
      active = null;
    },
    current() {
      return active;
    },
  };
}
