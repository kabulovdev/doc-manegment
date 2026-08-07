"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { FormEvent, useState } from "react";
import { authApi, ApiError } from "@/lib/api/client";
import { useAuthStore } from "@/lib/stores/auth-store";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Input } from "@/components/ui/v2/input";

export default function RegisterPage() {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      await authApi.register({ email, password, display_name: name });
      const res = await authApi.login({ email, password });
      setAuth(res.access_token, res.user);
      router.replace("/dashboard");
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="w-full max-w-sm" padding="lg">
      <div className="mb-4">
        <h1 className="text-lg font-semibold text-text">Create account</h1>
        <p className="text-[12px] text-text-2 mt-1">
          Start managing your documents.
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-3">
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">Name</label>
          <Input
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
          />
        </div>
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
            autoComplete="new-password"
            required
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <p className="text-[11px] text-text-3">At least 8 characters.</p>
        </div>
        {err && <p className="text-[12px] text-danger">{err}</p>}
        <Button
          type="submit"
          variant="accent"
          className="w-full"
          disabled={loading}
        >
          {loading ? "Creating…" : "Create account"}
        </Button>
      </form>

      <p className="mt-4 text-center text-[12px] text-text-3">
        Already have an account?{" "}
        <Link className="text-accent-2 hover:underline" href="/login">
          Sign in
        </Link>
      </p>
    </Card>
  );
}
