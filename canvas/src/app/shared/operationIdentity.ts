export type OperationId = object;

export interface OperationIdentity {
  begin(): OperationId;
  invalidate(): void;
  isCurrent(operation: OperationId): boolean;
}

export function createOperationIdentity(): OperationIdentity {
  let active: OperationId = {};
  return {
    begin() {
      active = {};
      return active;
    },
    invalidate() {
      active = {};
    },
    isCurrent(operation) {
      return active === operation;
    },
  };
}
