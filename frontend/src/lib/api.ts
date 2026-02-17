export type ApiEnvelope = {
  success?: boolean;
  data?: unknown;
  error?: string;
};

export type JsonRecord = Record<string, unknown>;

export function isRecord(value: unknown): value is JsonRecord {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

export function asString(value: unknown): string | null {
  return typeof value === "string" ? value : null;
}

export function asNumber(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

export function asBoolean(value: unknown): boolean | null {
  return typeof value === "boolean" ? value : null;
}

export function asArray(value: unknown): unknown[] | null {
  return Array.isArray(value) ? value : null;
}

export async function apiJson(
  apiBaseUrl: string,
  path: string,
  init?: RequestInit,
): Promise<{ ok: boolean; status: number; json: ApiEnvelope | null }> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    credentials: "include",
    ...init,
    headers: {
      ...(init?.headers ?? {}),
    },
  });

  const json = (await response.json().catch(() => null)) as ApiEnvelope | null;
  return { ok: response.ok, status: response.status, json };
}

