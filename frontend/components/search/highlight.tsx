import { ReactNode } from "react";

export function highlight(text: string, query: string): ReactNode {
  if (!query) return text;
  const q = query.toLowerCase();
  const t = text.toLowerCase();
  const nodes: ReactNode[] = [];
  let i = 0;
  while (i < text.length) {
    const idx = t.indexOf(q, i);
    if (idx < 0) {
      nodes.push(text.slice(i));
      break;
    }
    if (idx > i) nodes.push(text.slice(i, idx));
    nodes.push(
      <mark
        key={idx}
        className="bg-accent-soft text-accent-2 rounded px-0.5"
      >
        {text.slice(idx, idx + q.length)}
      </mark>,
    );
    i = idx + q.length;
  }
  return <>{nodes}</>;
}
