"use client";

import { ApiError, apiFetch } from "../api/client";
import { filesApi, InitUploadInput } from "../api/files";
import { useAuthStore } from "../stores/auth-store";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1";
const MAX_CONCURRENCY = 4;

export interface UploadProgress {
  loaded: number;
  total: number;
  percent: number;
}

export interface UploadResult {
  fileID: string;
}

interface CompletePart {
  part_number: number;
  etag: string;
}

export async function uploadFile(
  file: File,
  input: Omit<InitUploadInput, "name" | "size" | "mime"> & {
    name?: string;
    mime?: string;
  },
  onProgress?: (p: UploadProgress) => void,
  signal?: AbortSignal,
): Promise<UploadResult> {
  const init = await filesApi.initUpload({
    ...input,
    name: input.name ?? file.name,
    size: file.size,
    mime: input.mime ?? file.type ?? "application/octet-stream",
  });

  if (!init.multipart) {
    await putOrPost(
      "POST",
      `${API_BASE}/files/upload/${init.file_id}/single`,
      file,
      file.type || "application/octet-stream",
      signal,
      (loaded) =>
        onProgress?.({
          loaded,
          total: file.size,
          percent: file.size === 0 ? 0 : (loaded / file.size) * 100,
        }),
    );
    return { fileID: init.file_id };
  }

  try {
    const parts = await uploadAllParts(init.file_id, file, init.part_size, onProgress, signal);
    await apiFetch(`/files/upload/${init.file_id}/complete`, {
      method: "POST",
      body: JSON.stringify({ parts }),
    });
    return { fileID: init.file_id };
  } catch (err) {
    try {
      await filesApi.abortUpload(init.file_id);
    } catch {
      /* ignore */
    }
    throw err;
  }
}

async function uploadAllParts(
  fileID: string,
  file: File,
  partSize: number,
  onProgress?: (p: UploadProgress) => void,
  signal?: AbortSignal,
): Promise<CompletePart[]> {
  const total = file.size;
  const numParts = Math.ceil(total / partSize);
  const result = new Array<CompletePart>(numParts);
  const loaded = new Array<number>(numParts).fill(0);

  const report = () => {
    if (!onProgress) return;
    const sum = loaded.reduce((a, b) => a + b, 0);
    onProgress({ loaded: sum, total, percent: total === 0 ? 0 : (sum / total) * 100 });
  };

  let cursor = 0;

  async function worker() {
    while (cursor < numParts) {
      if (signal?.aborted) throw new DOMException("aborted", "AbortError");
      const i = cursor++;
      const start = i * partSize;
      const end = Math.min(start + partSize, total);
      const blob = file.slice(start, end);
      const partNumber = i + 1;
      const url = `${API_BASE}/files/upload/${fileID}/part?part_number=${partNumber}`;
      const text = await putOrPost("PUT", url, blob, "application/octet-stream", signal, (bytes) => {
        loaded[i] = bytes;
        report();
      });
      const parsed = JSON.parse(text) as { etag: string };
      result[i] = { part_number: partNumber, etag: parsed.etag };
    }
  }

  const workers = Array.from(
    { length: Math.min(MAX_CONCURRENCY, numParts) },
    () => worker(),
  );
  await Promise.all(workers);
  report();
  return result;
}

function putOrPost(
  method: "PUT" | "POST",
  url: string,
  body: Blob,
  mime: string,
  signal?: AbortSignal,
  onBytes?: (loaded: number) => void,
): Promise<string> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open(method, url);
    xhr.setRequestHeader("Content-Type", mime);
    const token = useAuthStore.getState().accessToken ?? "";
    if (token) xhr.setRequestHeader("Authorization", `Bearer ${token}`);
    xhr.withCredentials = true;
    if (onBytes) {
      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) onBytes(e.loaded);
      });
    }
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(xhr.responseText ?? "");
        return;
      }
      let parsed: { error?: string; message?: string } = {};
      try {
        parsed = JSON.parse(xhr.responseText);
      } catch {
        /* ignore */
      }
      reject(
        new ApiError(xhr.status, parsed.error ? { error: parsed.error, message: parsed.message } : { error: "http_error" }),
      );
    };
    xhr.onerror = () => reject(new ApiError(0, { error: "network_error" }));
    xhr.onabort = () => reject(new DOMException("aborted", "AbortError"));
    signal?.addEventListener("abort", () => xhr.abort());
    xhr.send(body);
  });
}
