export function domainSpec(content: string): unknown {
  try {
    const parsed = JSON.parse(content) as unknown;
    return typeof parsed === "object" && parsed !== null && "domain_spec" in parsed ? parsed.domain_spec : null;
  } catch {
    return null;
  }
}
