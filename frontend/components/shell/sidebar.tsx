"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { authApi } from "@/lib/api/client";
import { storagesApi } from "@/lib/api/storages";
import { useAuthStore } from "@/lib/stores/auth-store";
import { cn } from "@/lib/utils";
import { Avatar } from "@/components/ui/v2/avatar";
import { Icon } from "@/components/ui/v2/icon";
import { Progress } from "@/components/ui/v2/progress";
import { Tooltip } from "@/components/ui/v2/tooltip";

export interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  mobile?: boolean;
  onNavigate?: () => void;
}

interface NavItem {
  href: string;
  label: string;
  icon: string;
  exact?: boolean;
}

const primary: NavItem[] = [
  { href: "/dashboard", label: "Overview", icon: "LayoutDashboard", exact: true },
  { href: "/dashboard/files", label: "Files", icon: "Files" },
  { href: "/dashboard/storages", label: "Storages", icon: "Cloud" },
  { href: "/dashboard/tags", label: "Tags", icon: "Tags" },
  { href: "/dashboard/shares", label: "Shares", icon: "Share2" },
  { href: "/dashboard/search", label: "Search", icon: "Search" },
  { href: "/dashboard/ask", label: "Ask your docs", icon: "MessageSquare" },
  { href: "/dashboard/ai", label: "AI providers", icon: "Sparkles" },
];

const secondary: NavItem[] = [
  { href: "/dashboard/team", label: "Team", icon: "Users" },
  { href: "/dashboard/settings", label: "Settings", icon: "Settings" },
];

function isActive(pathname: string, item: NavItem) {
  if (item.exact) return pathname === item.href;
  return pathname === item.href || pathname.startsWith(`${item.href}/`);
}

function formatBytes(n: number) {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
}

function NavLink({
  item,
  active,
  collapsed,
  onClick,
}: {
  item: NavItem;
  active: boolean;
  collapsed: boolean;
  onClick?: () => void;
}) {
  const content = (
    <Link
      href={item.href}
      onClick={onClick}
      data-active={active || undefined}
      className={cn(
        "flex items-center h-8 gap-2.5 rounded text-[13px] transition-colors",
        collapsed ? "justify-center w-8 mx-auto" : "px-2.5",
        active
          ? "bg-accent-soft text-accent-2 font-medium"
          : "text-text-2 hover:bg-surface-2 hover:text-text",
      )}
    >
      <Icon name={item.icon} size={16} className={cn(active && "text-accent-2")} />
      {!collapsed && <span className="truncate">{item.label}</span>}
    </Link>
  );
  if (collapsed) {
    return (
      <Tooltip content={item.label} side="right">
        {content}
      </Tooltip>
    );
  }
  return content;
}

export function Sidebar({ collapsed, onToggle, mobile, onNavigate }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, clear } = useAuthStore();

  const { data: storages } = useQuery({
    queryKey: ["storages"],
    queryFn: () => storagesApi.list(),
    staleTime: 60_000,
  });

  const totalUsed = storages?.reduce((s, x) => s + x.used_bytes, 0) ?? 0;
  const storageCount = storages?.length ?? 0;

  async function onLogout() {
    try {
      await authApi.logout();
    } finally {
      clear();
      router.replace("/login");
    }
  }

  const showLabels = !collapsed || mobile;
  const effectiveCollapsed = collapsed && !mobile;

  return (
    <aside
      style={{
        width: effectiveCollapsed ? 56 : "var(--sidebar-w)",
      }}
      className={cn(
        "flex h-full flex-col border-r border-border bg-surface transition-[width] duration-200",
        mobile && "w-[var(--sidebar-w)] shadow-lg",
      )}
    >
      <div
        className={cn(
          "flex items-center border-b border-border",
          effectiveCollapsed ? "justify-center px-2" : "justify-between px-3",
        )}
        style={{ height: "var(--topbar-h)" }}
      >
        {showLabels ? (
          <Link href="/dashboard" className="flex items-center gap-2 min-w-0">
            <span className="inline-flex h-6 w-6 items-center justify-center rounded bg-accent text-white">
              <Icon name="Layers" size={14} />
            </span>
            <span className="text-[13px] font-semibold text-text truncate">
              Doc Manager
            </span>
          </Link>
        ) : (
          <Link href="/dashboard" aria-label="Doc Manager">
            <span className="inline-flex h-6 w-6 items-center justify-center rounded bg-accent text-white">
              <Icon name="Layers" size={14} />
            </span>
          </Link>
        )}
        {!mobile && (
          <button
            type="button"
            onClick={onToggle}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
            className={cn(
              "text-text-3 hover:text-text",
              effectiveCollapsed && "hidden",
            )}
          >
            <Icon name={collapsed ? "ChevronRight" : "ChevronLeft"} size={14} />
          </button>
        )}
      </div>

      {showLabels && (
        <div className="px-3 py-2 border-b border-border">
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded border border-border bg-surface-2 px-2 h-8 text-[12px] text-text-2 hover:text-text"
          >
            <span className="inline-flex h-4 w-4 items-center justify-center rounded bg-accent-soft text-accent-2 text-[10px] font-semibold">
              {(user?.display_name ?? "W").slice(0, 1).toUpperCase()}
            </span>
            <span className="flex-1 text-left truncate font-medium text-text">
              My Workspace
            </span>
            <Icon name="ChevronsUpDown" size={12} className="text-text-3" />
          </button>
        </div>
      )}

      <nav className={cn("flex-1 overflow-y-auto py-2", !effectiveCollapsed && "px-2")}>
        <div className="space-y-0.5">
          {primary.map((item) => (
            <NavLink
              key={item.href}
              item={item}
              active={isActive(pathname ?? "", item)}
              collapsed={effectiveCollapsed}
              onClick={onNavigate}
            />
          ))}
        </div>
        <div className={cn("my-3 h-px bg-border", effectiveCollapsed ? "mx-2" : "")} />
        <div className="space-y-0.5">
          {secondary.map((item) => (
            <NavLink
              key={item.href}
              item={item}
              active={isActive(pathname ?? "", item)}
              collapsed={effectiveCollapsed}
              onClick={onNavigate}
            />
          ))}
        </div>
      </nav>

      {showLabels && (
        <div className="px-3 py-2 border-t border-border">
          <div className="flex items-center justify-between text-[11px] text-text-3 mb-1">
            <span>Storage</span>
            <span className="font-mono">{formatBytes(totalUsed)}</span>
          </div>
          <Progress value={storageCount > 0 ? Math.min(totalUsed / (100 * 1024 ** 3), 1) * 100 : 0} />
          <div className="text-[11px] text-text-3 mt-1">
            {storageCount} storage{storageCount === 1 ? "" : "s"}
          </div>
        </div>
      )}

      <div
        className={cn(
          "flex items-center gap-2 border-t border-border",
          effectiveCollapsed ? "justify-center p-2" : "px-3 py-2",
        )}
      >
        {user ? (
          <Avatar name={user.display_name || user.email} size="sm" />
        ) : (
          <Avatar name="?" size="sm" />
        )}
        {showLabels && (
          <div className="min-w-0 flex-1">
            <div className="text-[12px] font-medium text-text truncate">
              {user?.display_name || "User"}
            </div>
            <div className="text-[11px] text-text-3 truncate">
              {user?.email}
            </div>
          </div>
        )}
        {showLabels ? (
          <button
            type="button"
            onClick={onLogout}
            aria-label="Log out"
            className="text-text-3 hover:text-text"
          >
            <Icon name="LogOut" size={14} />
          </button>
        ) : (
          <Tooltip content="Log out" side="right">
            <button
              type="button"
              onClick={onLogout}
              aria-label="Log out"
              className="text-text-3 hover:text-text"
            >
              <Icon name="LogOut" size={14} />
            </button>
          </Tooltip>
        )}
      </div>
    </aside>
  );
}
