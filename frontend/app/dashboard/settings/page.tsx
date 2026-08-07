"use client";

import { useState } from "react";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/v2/badge";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { Icon } from "@/components/ui/v2/icon";
import { Input } from "@/components/ui/v2/input";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Select } from "@/components/ui/v2/select";
import { Toggle } from "@/components/ui/v2/toggle";
import { ProfileTab } from "@/components/settings/profile-tab";
import { ApiTokensTab } from "@/components/settings/api-tokens-tab";

type TabKey =
  | "profile"
  | "workspace"
  | "security"
  | "api"
  | "billing"
  | "notifications";

const tabs: { key: TabKey; label: string; icon: string; soon?: boolean }[] = [
  { key: "profile", label: "Profile", icon: "User" },
  { key: "workspace", label: "Workspace", icon: "Building2", soon: true },
  { key: "security", label: "Security", icon: "Shield", soon: true },
  { key: "api", label: "API & Webhooks", icon: "Code" },
  { key: "billing", label: "Billing", icon: "CreditCard", soon: true },
  { key: "notifications", label: "Notifications", icon: "Bell", soon: true },
];

export default function SettingsPage() {
  usePageTitle("Settings", "Manage your account and workspace");
  const [tab, setTab] = useState<TabKey>("profile");

  return (
    <div className="space-y-6">
      <SectionHead level="h1" title="Settings" />
      <div className="flex flex-col gap-5 md:flex-row md:gap-6">
        <nav className="md:w-[220px] md:shrink-0">
          <ul className="space-y-0.5">
            {tabs.map((t) => {
              const active = t.key === tab;
              return (
                <li key={t.key}>
                  <button
                    type="button"
                    onClick={() => setTab(t.key)}
                    className={cn(
                      "flex items-center gap-2 w-full rounded px-2.5 h-8 text-[13px] transition-colors",
                      active
                        ? "bg-accent-soft text-accent-2 font-medium"
                        : "text-text-2 hover:bg-surface-2 hover:text-text",
                    )}
                  >
                    <Icon name={t.icon} size={14} />
                    <span className="flex-1 text-left truncate">{t.label}</span>
                    {t.soon && (
                      <Badge color="slate" className="text-[9px]">
                        Soon
                      </Badge>
                    )}
                  </button>
                </li>
              );
            })}
          </ul>
        </nav>

        <div className="flex-1 min-w-0">
          {tab === "profile" && <ProfileTab />}
          {tab === "workspace" && <WorkspaceTab />}
          {tab === "security" && <SecurityTab />}
          {tab === "api" && <ApiTab />}
          {tab === "billing" && <BillingTab />}
          {tab === "notifications" && <NotificationsTab />}
        </div>
      </div>
    </div>
  );
}

function WorkspaceTab() {
  return (
    <div className="space-y-4">
      <Card title="Workspace">
        <div className="space-y-3">
          <div className="space-y-1.5">
            <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Workspace name
            </div>
            <Input value="My Workspace" readOnly className="opacity-80" />
          </div>
          <div className="space-y-1.5">
            <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Shareable URL
            </div>
            <Input value="—" readOnly className="opacity-80 font-mono text-[11px]" />
          </div>
          <div className="space-y-1.5">
            <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Default upload storage
            </div>
            <Select disabled>
              <option>Auto (first connected)</option>
            </Select>
          </div>
        </div>
      </Card>
      <Card>
        <EmptyState
          icon={<Icon name="Building2" size={18} />}
          title="Workspace management coming soon"
          description="Rename your workspace, invite teammates, set defaults, and more."
        />
      </Card>
    </div>
  );
}

function SecurityTab() {
  return (
    <div className="space-y-4">
      <Card title="Password">
        <div className="space-y-3 opacity-70 pointer-events-none">
          <div className="space-y-1.5">
            <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Current password
            </div>
            <Input type="password" readOnly />
          </div>
          <div className="space-y-1.5">
            <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              New password
            </div>
            <Input type="password" readOnly />
          </div>
        </div>
      </Card>
      <Card
        title="Two-factor authentication"
        action={<Badge color="slate">Disabled</Badge>}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="text-[13px] text-text">Authenticator app</div>
            <div className="text-[11px] text-text-3">
              Require a one-time code on every sign-in.
            </div>
          </div>
          <Toggle checked={false} disabled aria-label="Enable 2FA" />
        </div>
      </Card>
      <Card>
        <EmptyState
          icon={<Icon name="Shield" size={18} />}
          title="Full security controls in the next release"
          description="Password changes, 2FA, and active sessions arrive soon."
        />
      </Card>
    </div>
  );
}

function ApiTab() {
  return <ApiTokensTab />;
}

function BillingTab() {
  return (
    <Card>
      <EmptyState
        icon={<Icon name="CreditCard" size={18} />}
        title="Billing will be available once subscriptions launch"
      />
    </Card>
  );
}

function NotificationsTab() {
  return (
    <Card>
      <EmptyState
        icon={<Icon name="Bell" size={18} />}
        title="Notification preferences coming soon"
        description="Email, in-app, and mobile notification controls."
      />
    </Card>
  );
}
