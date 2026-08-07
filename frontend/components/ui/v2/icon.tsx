import * as LucideIcons from "lucide-react";
import { cn } from "@/lib/utils";

type LucideComponent = (typeof LucideIcons)[keyof typeof LucideIcons];

export type IconName = keyof typeof LucideIcons;

const aliasMap: Record<string, string> = {
  search: "Search",
  settings: "Settings",
  sun: "Sun",
  moon: "Moon",
  gear: "Settings",
  close: "X",
  x: "X",
  plus: "Plus",
  minus: "Minus",
  check: "Check",
  chevronDown: "ChevronDown",
  chevronRight: "ChevronRight",
  chevronLeft: "ChevronLeft",
  chevronUp: "ChevronUp",
  menu: "Menu",
  more: "MoreHorizontal",
  moreVertical: "MoreVertical",
  upload: "Upload",
  download: "Download",
  trash: "Trash2",
  edit: "Pencil",
  folder: "Folder",
  file: "File",
  share: "Share2",
  link: "Link",
  tag: "Tag",
  star: "Star",
  bell: "Bell",
  user: "User",
  users: "Users",
  lock: "Lock",
  unlock: "Unlock",
  eye: "Eye",
  eyeOff: "EyeOff",
  home: "Home",
  inbox: "Inbox",
  cloud: "Cloud",
  refresh: "RefreshCw",
  filter: "Filter",
  grid: "LayoutGrid",
  list: "List",
  calendar: "Calendar",
  clock: "Clock",
  copy: "Copy",
  info: "Info",
  warning: "AlertTriangle",
  error: "AlertCircle",
  success: "CheckCircle2",
  sparkles: "Sparkles",
  command: "Command",
  arrowLeft: "ArrowLeft",
  arrowRight: "ArrowRight",
  arrowUp: "ArrowUp",
  arrowDown: "ArrowDown",
};

export function getIcon(name: string): LucideComponent | null {
  const mapped = aliasMap[name] ?? name;
  const Comp = (LucideIcons as Record<string, unknown>)[mapped];
  return (Comp as LucideComponent) ?? null;
}

export interface IconProps extends React.SVGAttributes<SVGSVGElement> {
  name: string;
  size?: number;
}

export function Icon({ name, size = 16, className, ...rest }: IconProps) {
  const Comp = getIcon(name) as React.ComponentType<React.SVGAttributes<SVGSVGElement> & { size?: number }> | null;
  if (!Comp) return null;
  return <Comp size={size} className={cn("shrink-0", className)} {...rest} />;
}
