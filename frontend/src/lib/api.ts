import type { ApiErrorBody } from "./types";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const API = `${BASE_URL}/api/v1`;

// ApiError carries the backend's error code so callers can branch on it
// (e.g. ALREADY_CLAIMED) instead of parsing messages.
export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API}${path}`, {
      // Always send the httpOnly player cookie across origins (Vercel↔Railway).
      credentials: "include",
      ...init,
    });
  } catch {
    throw new ApiError(0, "NETWORK", "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const body = await res.json().catch(() => null);
  if (!res.ok) {
    const err = (body as ApiErrorBody | null)?.error;
    throw new ApiError(
      res.status,
      err?.code ?? "UNKNOWN",
      err?.message ?? "เกิดข้อผิดพลาด",
    );
  }
  return body as T;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, data?: unknown) =>
    request<T>(path, {
      method: "POST",
      headers: data !== undefined ? { "Content-Type": "application/json" } : undefined,
      body: data !== undefined ? JSON.stringify(data) : undefined,
    }),
};

// Shared SWR fetcher for GET endpoints.
export const fetcher = <T>(path: string) => api.get<T>(path);
