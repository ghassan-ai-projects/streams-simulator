interface Foo {
  readonly a: number;
}
export function x(value: unknown): Foo {
  return value as unknown as Foo;
}
