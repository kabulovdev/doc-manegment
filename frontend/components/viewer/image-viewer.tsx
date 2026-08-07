"use client";

import { useEffect } from "react";

interface Props {
  url: string;
  alt: string;
  withCredentials?: boolean;
}

export function ImageViewer({ url, alt, withCredentials }: Props) {
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
      className="flex justify-center select-none"
      onContextMenu={(e) => e.preventDefault()}
    >
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={url}
        alt={alt}
        draggable={false}
        crossOrigin={withCredentials ? "use-credentials" : undefined}
        className="max-h-[85vh] max-w-full rounded-lg shadow-sm"
        style={{ pointerEvents: "auto", userSelect: "none" }}
      />
    </div>
  );
}
