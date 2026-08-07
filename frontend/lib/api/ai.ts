"use client";

import { apiFetch } from "./client";

export type AIProviderKind =
  | "anthropic"
  | "openai"
  | "google"
  | "openai-compatible"
  | "ollama";

export type AICapability = "chat" | "embed" | "transcribe";

export interface AIConfig {
  id: string;
  display_name: string;
  provider: AIProviderKind;
  base_url?: string;
  key_prefix?: string;
  chat_model?: string;
  embed_model?: string;
  transcribe_model?: string;
  capabilities: AICapability[];
  is_default_chat: boolean;
  is_default_embed: boolean;
  is_default_transcribe: boolean;
  used_tokens_in: number;
  used_tokens_out: number;
  last_used_at?: string | null;
  last_tested_at?: string | null;
  last_error?: string;
  created_at: string;
}

export interface CreateAIConfigInput {
  display_name: string;
  provider: AIProviderKind;
  base_url?: string;
  api_key?: string;
  chat_model?: string;
  embed_model?: string;
  transcribe_model?: string;
  capabilities?: AICapability[];
  default_chat?: boolean;
  default_embed?: boolean;
  default_transcribe?: boolean;
}

export interface UpdateAIConfigInput {
  display_name?: string;
  chat_model?: string;
  embed_model?: string;
  transcribe_model?: string;
  capabilities?: AICapability[];
}

export interface AIUsageTotals {
  ai_config_id: string;
  tokens_in: number;
  tokens_out: number;
  calls: number;
}

export interface SuggestedField {
  key: string;
  value: string;
  type: "text" | "number" | "date" | "boolean";
  confidence?: number;
}

export interface SuggestedTag {
  name: string;
  confidence?: number;
}

export type ExtractionStatus = "pending" | "ready" | "unsupported" | "error";

export interface AISuggestionsMeta {
  provider?: string;
  model?: string;
  tokens_in?: number;
  tokens_out?: number;
  latency_ms?: number;
  extraction_status: ExtractionStatus;
}

export interface SuggestFieldsResponse {
  fields: SuggestedField[];
  meta: AISuggestionsMeta;
}

export interface SuggestTagsResponse {
  tags: SuggestedTag[];
  meta: AISuggestionsMeta;
}

export interface SuggestNameResponse {
  name: string;
  meta: AISuggestionsMeta;
}

export interface SemanticSearchHit {
  file_id: string;
  score: number;
  name: string;
  mime?: string;
  folder_id?: string;
  text_preview?: string;
}

export interface SemanticSearchResponse {
  hits: SemanticSearchHit[];
}

export interface AskCitation {
  index: number;
  file_id: string;
  name: string;
  mime?: string;
  score?: number;
  snippet?: string;
}

export interface AskResponse {
  answer: string;
  citations: AskCitation[];
  provider?: string;
  model?: string;
  tokens_in?: number;
  tokens_out?: number;
  latency_ms?: number;
}

export interface SuggestFolderResponse {
  folder_id?: string;
  folder_name?: string;
  score?: number;
}

export type AIExtractionStatus = "none" | "pending" | "ready" | "failed";

export interface AIExtraction {
  status: AIExtractionStatus;
  text?: string;
  model?: string;
  provider?: string;
  provider_id?: string;
  tokens_in?: number;
  tokens_out?: number;
  extracted_at?: string;
  error?: string;
}

export interface ProcessDocumentInput {
  provider_id?: string;
}

export const aiApi = {
  list: () => apiFetch<AIConfig[]>("/ai/configs"),
  get: (id: string) => apiFetch<AIConfig>(`/ai/configs/${id}`),
  create: (input: CreateAIConfigInput) =>
    apiFetch<AIConfig>("/ai/configs", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  update: (id: string, input: UpdateAIConfigInput) =>
    apiFetch<AIConfig>(`/ai/configs/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    }),
  remove: (id: string) =>
    apiFetch<void>(`/ai/configs/${id}`, { method: "DELETE" }),
  test: (id: string) =>
    apiFetch<{ ok: boolean; error?: string }>(`/ai/configs/${id}/test`, {
      method: "POST",
    }),
  setDefault: (id: string, capability: AICapability) =>
    apiFetch<AIConfig>(`/ai/configs/${id}/default`, {
      method: "POST",
      body: JSON.stringify({ capability }),
    }),
  usage: () => apiFetch<AIUsageTotals[]>("/ai/usage"),
  suggestFields: (fileID: string) =>
    apiFetch<SuggestFieldsResponse>(`/files/${fileID}/ai/extract-fields`, {
      method: "POST",
    }),
  suggestTags: (fileID: string) =>
    apiFetch<SuggestTagsResponse>(`/files/${fileID}/ai/suggest-tags`, {
      method: "POST",
    }),
  suggestName: (fileID: string) =>
    apiFetch<SuggestNameResponse>(`/files/${fileID}/ai/suggest-name`, {
      method: "POST",
    }),
  suggestFolder: (fileID: string) =>
    apiFetch<SuggestFolderResponse>(`/files/${fileID}/ai/suggest-folder`, {
      method: "POST",
    }),
  semanticSearch: (query: string, limit = 10) =>
    apiFetch<SemanticSearchResponse>("/ai/search", {
      method: "POST",
      body: JSON.stringify({ query, limit }),
    }),
  ask: (question: string, limit = 5) =>
    apiFetch<AskResponse>("/ai/ask", {
      method: "POST",
      body: JSON.stringify({ question, limit }),
    }),
  processDocument: (fileID: string, input: ProcessDocumentInput = {}) =>
    apiFetch<AIExtraction>(`/files/${fileID}/ai/process`, {
      method: "POST",
      body: JSON.stringify(input),
    }),
  getExtraction: (fileID: string) =>
    apiFetch<AIExtraction>(`/files/${fileID}/ai/extraction`),
  fileChat: (fileID: string, question: string) =>
    apiFetch<AskResponse>(`/files/${fileID}/ai/chat`, {
      method: "POST",
      body: JSON.stringify({ question }),
    }),
};
