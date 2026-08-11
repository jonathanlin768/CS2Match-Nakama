export function toStringValue(value: unknown): string {
  return value === null || value === undefined || Array.isArray(value) ? '' : String(value)
}

export function toNumberValue(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export function toListValue(value: unknown): string[] {
  if (Array.isArray(value)) return value.map(String)
  return value === null || value === undefined || value === '' ? [] : String(value).split(',').map((item) => item.trim()).filter(Boolean)
}
