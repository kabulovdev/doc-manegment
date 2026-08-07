"use client";

import { useEffect, useMemo, useState } from "react";
import { Document, Page, pdfjs } from "react-pdf";
import "react-pdf/dist/Page/AnnotationLayer.css";
import "react-pdf/dist/Page/TextLayer.css";
import { ChevronLeft, ChevronRight, ZoomIn, ZoomOut } from "lucide-react";

// Worker is copied to public/pdf.worker.min.mjs at build time so the exact
// pdfjs-dist version we shipped is served — no CDN mismatch or 404s.
pdfjs.GlobalWorkerOptions.workerSrc = "/pdf.worker.min.mjs";

type FileSource = string | Blob;

interface Props {
  // Either a URL string, a Blob, or a prepared react-pdf file object.
  file?: FileSource;
  // Legacy URL prop (kept for backward compat with share page).
  url?: string;
  withCredentials?: boolean;
  httpHeaders?: Record<string, string>;
}

export function PDFViewer({ file, url, withCredentials, httpHeaders }: Props) {
  const [pageNum, setPageNum] = useState(1);
  const [numPages, setNumPages] = useState(0);
  const [scale, setScale] = useState(1.2);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const fileProp = useMemo(() => {
    const src = file ?? url;
    if (!src) return null;
    if (typeof src === "string") {
      if (withCredentials || httpHeaders) {
        return { url: src, withCredentials: !!withCredentials, httpHeaders: httpHeaders ?? {} };
      }
      return src;
    }
    return src; // Blob — react-pdf accepts it directly
  }, [file, url, withCredentials, httpHeaders]);

  useEffect(() => {
    const block = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && ["s", "p", "S", "P"].includes(e.key)) {
        e.preventDefault();
      }
    };
    window.addEventListener("keydown", block);
    return () => window.removeEventListener("keydown", block);
  }, []);

  return (
    <div
      className="flex flex-col items-center gap-4"
      onContextMenu={(e) => e.preventDefault()}
    >
      <div className="flex items-center gap-2 rounded-lg bg-white border border-slate-200 px-3 py-1.5 shadow-sm">
        <button
          onClick={() => setPageNum((p) => Math.max(1, p - 1))}
          disabled={pageNum <= 1}
          className="text-slate-500 hover:text-slate-900 disabled:opacity-30"
        >
          <ChevronLeft size={16} />
        </button>
        <span className="text-xs tabular-nums px-2">
          {pageNum} / {numPages || "…"}
        </span>
        <button
          onClick={() => setPageNum((p) => Math.min(numPages, p + 1))}
          disabled={pageNum >= numPages}
          className="text-slate-500 hover:text-slate-900 disabled:opacity-30"
        >
          <ChevronRight size={16} />
        </button>
        <div className="mx-1 h-5 w-px bg-slate-200" />
        <button
          onClick={() => setScale((s) => Math.max(0.5, s - 0.2))}
          className="text-slate-500 hover:text-slate-900"
        >
          <ZoomOut size={16} />
        </button>
        <span className="text-xs tabular-nums px-1">{Math.round(scale * 100)}%</span>
        <button
          onClick={() => setScale((s) => Math.min(3, s + 0.2))}
          className="text-slate-500 hover:text-slate-900"
        >
          <ZoomIn size={16} />
        </button>
      </div>
      <div className="select-none w-full flex justify-center">
        {fileProp ? (
          <Document
            file={fileProp}
            onLoadSuccess={(d) => {
              setNumPages(d.numPages);
              setLoadErr(null);
            }}
            onLoadError={(err) => {
              // eslint-disable-next-line no-console
              console.error("[PDFViewer] load error", err);
              setLoadErr(err?.message ?? String(err));
            }}
            loading={<div className="text-sm text-slate-500">Loading PDF…</div>}
            error={
              <div className="text-sm text-red-600">
                Failed to load PDF.
                {loadErr && (
                  <div className="mt-2 text-xs font-mono text-red-500 max-w-xl break-all">
                    {loadErr}
                  </div>
                )}
              </div>
            }
          >
            <Page
              pageNumber={pageNum}
              scale={scale}
              renderTextLayer
              renderAnnotationLayer={false}
            />
          </Document>
        ) : (
          <div className="text-sm text-slate-500">No file source.</div>
        )}
      </div>
    </div>
  );
}
