"use client";

import { useParams } from "next/navigation";
import { FilesView } from "@/components/files/files-view";

export default function FilesFolderPage() {
  const params = useParams<{ folderId: string }>();
  return <FilesView folderID={params.folderId} />;
}
