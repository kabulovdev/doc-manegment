"use client";

import { Dialog } from "@/components/ui/v2/dialog";
import { Button } from "@/components/ui/v2/button";

export interface UploadDialogStubProps {
  open: boolean;
  onClose: () => void;
}

export function UploadDialogStub({ open, onClose }: UploadDialogStubProps) {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Upload files"
      description="The full upload experience arrives in Phase C."
      footer={
        <Button variant="secondary" onClick={onClose}>
          Close
        </Button>
      }
    >
      <div className="flex items-center justify-center rounded border border-dashed border-border bg-surface-2 py-10 text-[12px] text-text-3">
        Drop files here or pick a storage (coming soon).
      </div>
    </Dialog>
  );
}
