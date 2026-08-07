"use client";

import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { aiApi, AIConfig } from "@/lib/api/ai";
import { ApiError } from "@/lib/api/client";
import { Button } from "@/components/ui/v2/button";
import { Dialog } from "@/components/ui/v2/dialog";
import { Icon } from "@/components/ui/v2/icon";
import { Select } from "@/components/ui/v2/select";

interface Props {
  open: boolean;
  fileID: string | null;
  fileName?: string;
  onClose: () => void;
  onDone?: (result: { processed: boolean }) => void;
}

// AiProcessModal is shown right after an upload completes. It asks the user
// whether they want to run a one-shot vision-AI extraction on the newly
// uploaded file so downstream tools (Fields / Tags / Rename / Folder /
// Ask-your-docs) can work against clean text instead of re-reading the raw
// bytes every time.
export function AiProcessModal({
  open,
  fileID,
  fileName,
  onClose,
  onDone,
}: Props) {
  const { data: configs = [] } = useQuery<AIConfig[]>({
    queryKey: ["ai-configs"],
    queryFn: () => aiApi.list(),
    enabled: open,
    staleTime: 60_000,
  });

  const chatConfigs = useMemo(
    () => configs.filter((c) => c.capabilities.includes("chat")),
    [configs],
  );
  const defaultConfig = useMemo(
    () =>
      chatConfigs.find((c) => c.is_default_chat) ?? chatConfigs[0] ?? null,
    [chatConfigs],
  );
  const [providerID, setProviderID] = useState<string>("");

  useEffect(() => {
    if (open && defaultConfig) setProviderID(defaultConfig.id);
  }, [open, defaultConfig]);

  const process = useMutation({
    mutationFn: () => {
      if (!fileID) throw new Error("Missing file id");
      return aiApi.processDocument(fileID, {
        provider_id: providerID || undefined,
      });
    },
    onSuccess: () => {
      onDone?.({ processed: true });
      onClose();
    },
  });

  const hasProvider = chatConfigs.length > 0;
  const loading = process.isPending;
  const errMsg =
    process.error instanceof ApiError
      ? process.error.message
      : process.error instanceof Error
        ? process.error.message
        : null;

  return (
    <Dialog
      open={open}
      onClose={() => {
        if (!loading) onClose();
      }}
      title={
        <span className="flex items-center gap-2">
          <Icon name="Sparkles" size={14} className="text-accent-2" />
          Tahlil qilasizmi?
        </span>
      }
      description={
        fileName ? (
          <span className="text-text-3">
            Fayl: <span className="text-text-2">{fileName}</span>
          </span>
        ) : undefined
      }
      size="md"
      footer={
        <>
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              onDone?.({ processed: false });
              onClose();
            }}
            disabled={loading}
          >
            Keyinroq
          </Button>
          <Button
            type="button"
            variant="accent"
            onClick={() => process.mutate()}
            disabled={loading || !hasProvider || !fileID}
          >
            {loading ? "Tahlil qilinmoqda…" : "Tahlil qilish"}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <div className="rounded-lg bg-accent-soft/40 border border-accent-border px-3 py-2.5 text-[12px] text-text-2">
          Hujjatni bir marta AI ga yuborib ichidagi barcha matn va
          maydonlarni olib qo&rsquo;yamiz. Shundan keyin{" "}
          <span className="font-medium text-text">
            Fields / Tags / Rename / Folder
          </span>{" "}
          tugmalari shu cached matn bilan tezroq va arzonroq ishlaydi.
        </div>

        {!hasProvider ? (
          <div className="rounded border border-red-200 bg-red-50 px-3 py-2 text-[12px] text-red-700">
            Chat qobiliyatli AI provayder yo&rsquo;q. AI providers sahifasida
            bittasini qo&rsquo;shing.
          </div>
        ) : (
          <div className="space-y-1.5">
            <label className="block text-[12px] font-medium text-text-2">
              Provayder
            </label>
            <Select
              value={providerID}
              onChange={(e) => setProviderID(e.target.value)}
              disabled={loading}
            >
              {chatConfigs.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.display_name} · {c.provider}
                  {c.chat_model ? ` · ${c.chat_model}` : ""}
                </option>
              ))}
            </Select>
          </div>
        )}

        {errMsg && (
          <div className="rounded border border-red-200 bg-red-50 px-3 py-2 text-[12px] text-red-700 break-words">
            {errMsg}
          </div>
        )}
      </div>
    </Dialog>
  );
}
