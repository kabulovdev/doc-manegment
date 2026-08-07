"use client";

import { useParams } from "next/navigation";
import { FileDetail } from "@/components/files/file-detail";

export default function FileViewPage() {
  const params = useParams<{ fileId: string }>();
  return <FileDetail fileID={params.fileId} />;
}
