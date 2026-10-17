export async function x() {
  const secret = await import("../app/internal/secret");
  return secret;
}
