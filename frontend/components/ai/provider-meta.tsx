import { AICapability, AIProviderKind } from "@/lib/api/ai";

export interface ProviderMeta {
  kind: AIProviderKind;
  label: string;
  tagline: string;
  color: string;
  defaultCapabilities: AICapability[];
  hasChat: boolean;
  hasEmbed: boolean;
  hasTranscribe: boolean;
  chatDefault: string;
  embedDefault?: string;
  transcribeDefault?: string;
  baseUrlHint?: string;
  keyHint: string;
  docsUrl?: string;
}

export const providerMeta: Record<AIProviderKind, ProviderMeta> = {
  anthropic: {
    kind: "anthropic",
    label: "Anthropic",
    tagline: "Claude family — best-in-class for chat",
    color: "#d97706",
    defaultCapabilities: ["chat"],
    hasChat: true,
    hasEmbed: false,
    hasTranscribe: false,
    chatDefault: "claude-haiku-4-5",
    keyHint: "sk-ant-…",
    docsUrl: "https://console.anthropic.com/settings/keys",
  },
  openai: {
    kind: "openai",
    label: "OpenAI",
    tagline: "Chat, embeddings, and Whisper transcription",
    color: "#10b981",
    defaultCapabilities: ["chat", "embed", "transcribe"],
    hasChat: true,
    hasEmbed: true,
    hasTranscribe: true,
    chatDefault: "gpt-4o-mini",
    embedDefault: "text-embedding-3-small",
    transcribeDefault: "whisper-1",
    keyHint: "sk-…",
    docsUrl: "https://platform.openai.com/api-keys",
  },
  google: {
    kind: "google",
    label: "Google Gemini",
    tagline: "Gemini 2.0 Flash + text-embedding-004",
    color: "#6366f1",
    defaultCapabilities: ["chat", "embed"],
    hasChat: true,
    hasEmbed: true,
    hasTranscribe: false,
    chatDefault: "gemini-2.0-flash",
    embedDefault: "text-embedding-004",
    keyHint: "AIza…",
    docsUrl: "https://aistudio.google.com/app/apikey",
  },
  "openai-compatible": {
    kind: "openai-compatible",
    label: "OpenAI-compatible",
    tagline: "OpenRouter, Groq, Together, vLLM, LM Studio",
    color: "#8b5cf6",
    defaultCapabilities: ["chat"],
    hasChat: true,
    hasEmbed: true,
    hasTranscribe: false,
    chatDefault: "",
    baseUrlHint: "https://openrouter.ai/api",
    keyHint: "sk-or-…",
    docsUrl: "https://openrouter.ai/keys",
  },
  ollama: {
    kind: "ollama",
    label: "Ollama (self-hosted)",
    tagline: "Run models locally — zero per-token cost",
    color: "#64748b",
    defaultCapabilities: ["chat", "embed"],
    hasChat: true,
    hasEmbed: true,
    hasTranscribe: false,
    chatDefault: "llama3.2",
    embedDefault: "nomic-embed-text",
    baseUrlHint: "http://ollama:11434",
    keyHint: "(leave empty)",
    docsUrl: "https://ollama.com",
  },
};

export const providerList: ProviderMeta[] = [
  providerMeta.anthropic,
  providerMeta.openai,
  providerMeta.google,
  providerMeta["openai-compatible"],
  providerMeta.ollama,
];
