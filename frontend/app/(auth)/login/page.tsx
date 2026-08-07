"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { FormEvent, useState } from "react";
import { authApi, ApiError } from "@/lib/api/client";
import { useAuthStore } from "@/lib/stores/auth-store";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Input } from "@/components/ui/v2/input";

export default function LoginPage() {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      const res = await authApi.login({ email, password });
      setAuth(res.access_token, res.user);
      router.replace("/dashboard");
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="w-full max-w-sm" padding="lg">
      <div className="mb-4">
        <h1 className="text-lg font-semibold text-text">Sign in</h1>
        <p className="text-[12px] text-text-2 mt-1">Welcome back.</p>
      </div>

      <form onSubmit={onSubmit} className="space-y-3">
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">Email</label>
          <Input
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">Password</label>
          <Input
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        {err && <p className="text-[12px] text-danger">{err}</p>}
        <Button
          type="submit"
          variant="accent"
          className="w-full"
          disabled={loading}
        >
          {loading ? "Signing in…" : "Sign in"}
        </Button>
      </form>

      <p className="mt-4 text-center text-[12px] text-text-3">
        No account?{" "}
        <Link className="text-accent-2 hover:underline" href="/register">
          Register
        </Link>
      </p>
    </Card>
  );
}
