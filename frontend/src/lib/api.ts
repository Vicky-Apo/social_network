const DEFAULT_API_BASE_URL = "http://localhost:8080";

export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") || DEFAULT_API_BASE_URL;
}

export function toApiUrl(path: string, baseUrl: string = getApiBaseUrl()): string {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${normalized}`;
}

export async function apiFetch(
  path: string,
  options: RequestInit = {},
  baseUrl: string = getApiBaseUrl(),
): Promise<Response> {
  return fetch(toApiUrl(path, baseUrl), {
    credentials: options.credentials ?? "include",
    ...options,
  });
}

export async function apiFetchJson<T>(
  path: string,
  options: RequestInit = {},
  baseUrl: string = getApiBaseUrl(),
): Promise<{ response: Response; result: T | null }> {
  const response = await apiFetch(path, options, baseUrl);
  const result = (await response.json().catch(() => null)) as T | null;
  return { response, result };
}
