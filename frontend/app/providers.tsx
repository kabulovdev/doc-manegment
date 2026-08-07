"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { authApi } from "@/lib/api/client";
import { useAuthStore } from "@/lib/stores/auth-store";
import { TweaksProvider } from "@/lib/tweaks/provider";

export function Providers({ children }: { children: React.ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { staleTime: 30_000, retry: 1 },
        },
      }),
  );

  const setAuth = useAuthStore((s) => s.setAuth);
  const [bootstrapped, setBootstrapped] = useState(false);

  useEffect(() => {
    (async () => {
      const ok = await authApi.refresh();
      if (ok) {
        try {
          const user = await authApi.me();
          const token = useAuthStore.getState().accessToken;
          if (token) setAuth(token, user);
        } catch {
          /* no-op */
        }
      }
      setBootstrapped(true);
    })();
  }, [setAuth]);

  if (!bootstrapped) {
    return (
      <TweaksProvider>
        <div className="flex min-h-screen items-center justify-center text-sm text-text-2">
          Loading…
        </div>
      </TweaksProvider>
    );
  }

  return (
    <TweaksProvider>
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    </TweaksProvider>
  );
}
