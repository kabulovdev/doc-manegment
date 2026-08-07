import { cn } from "@/lib/utils";

export type FileKind =
  | "pdf"
  | "image"
  | "excel"
  | "word"
  | "powerpoint"
  | "archive"
  | "video"
  | "audio"
  | "code"
  | "text"
  | "other";

const kindMap: Record<FileKind, { bg: string; text: string; label: string }> = {
  pdf: { bg: "bg-red-50 border-red-200", text: "text-red-700", label: "PDF" },
  image: { bg: "bg-sky-50 border-sky-200", text: "text-sky-700", label: "IMG" },
  excel: { bg: "bg-emerald-50 border-emerald-200", text: "text-emerald-700", label: "XLS" },
  word: { bg: "bg-blue-50 border-blue-200", text: "text-blue-700", label: "DOC" },
  powerpoint: { bg: "bg-orange-50 border-orange-200", text: "text-orange-700", label: "PPT" },
  archive: { bg: "bg-amber-50 border-amber-200", text: "text-amber-700", label: "ZIP" },
  video: { bg: "bg-violet-50 border-violet-200", text: "text-violet-700", label: "VID" },
  audio: { bg: "bg-pink-50 border-pink-200", text: "text-pink-700", label: "AUD" },
  code: { bg: "bg-slate-100 border-slate-200", text: "text-slate-700", label: "CODE" },
  text: { bg: "bg-surface-2 border-border", text: "text-text-2", label: "TXT" },
  other: { bg: "bg-surface-2 border-border", text: "text-text-2", label: "FILE" },
};

export function kindFromMime(mime?: string | null, name?: string): FileKind {
  const m = (mime ?? "").toLowerCase();
  const ext = (name?.split(".").pop() ?? "").toLowerCase();
  if (m.includes("pdf") || ext === "pdf") return "pdf";
  if (m.startsWith("image/") || ["png", "jpg", "jpeg", "gif", "webp", "svg"].includes(ext)) return "image";
  if (m.includes("spreadsheet") || m.includes("excel") || ["xls", "xlsx", "csv"].includes(ext)) return "excel";
  if (m.includes("word") || ["doc", "docx"].includes(ext)) return "word";
  if (m.includes("presentation") || ["ppt", "pptx"].includes(ext)) return "powerpoint";
  if (m.includes("zip") || m.includes("compress") || ["zip", "tar", "gz", "rar", "7z"].includes(ext)) return "archive";
  if (m.startsWith("video/") || ["mp4", "mov", "webm", "mkv"].includes(ext)) return "video";
  if (m.startsWith("audio/") || ["mp3", "wav", "flac"].includes(ext)) return "audio";
  if (["ts", "tsx", "js", "jsx", "py", "go", "rs", "java", "c", "cpp", "json", "yml", "yaml"].includes(ext)) return "code";
  if (m.startsWith("text/") || ["txt", "md"].includes(ext)) return "text";
  return "other";
}

export interface FileIconProps {
  mime?: string | null;
  name?: string;
  size?: "sm" | "md" | "lg";
  className?: string;
}

const sizeMap = {
  sm: "h-6 w-8 text-[9px]",
  md: "h-8 w-10 text-[10px]",
  lg: "h-10 w-12 text-[11px]",
};

export function FileIcon({ mime, name, size = "md", className }: FileIconProps) {
  const kind = kindFromMime(mime, name);
  const c = kindMap[kind];
  return (
    <span
      className={cn(
        "inline-flex items-center justify-center rounded border font-semibold uppercase tracking-wide",
        c.bg,
        c.text,
        sizeMap[size],
        className,
      )}
    >
      {c.label}
    </span>
  );
}
