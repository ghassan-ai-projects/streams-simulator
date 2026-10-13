export function env(name: string): string {
  const value = process.env[name];
  if (value === undefined) {
    throw new Error(`${name} is not set; run the tests through playwright so global-setup starts the bridges`);
  }
  return value;
}
