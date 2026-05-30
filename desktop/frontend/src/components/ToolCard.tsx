import { CodeViewer } from "./CodeViewer";
import { DiffView } from "./DiffView";
import type { Item } from "../lib/useController";

type ToolItem = Extract<Item, { kind: "tool" }>;

const STATUS_LABEL: Record<ToolItem["status"], string> = {
  running: "running…",
  done: "done",
  error: "error",
};

// editDiff extracts before/after from an edit_file call's args so the diff seam
// can render it. Args arrive complete on dispatch for the kernel; a partial /
// non-JSON value just yields null and we fall back to showing the raw args.
function editDiff(name: string, args: string): { original: string; modified: string } | null {
  if (name !== "edit_file") return null;
  try {
    const a = JSON.parse(args) as { old_string?: string; new_string?: string };
    if (typeof a.old_string === "string" && typeof a.new_string === "string") {
      return { original: a.old_string, modified: a.new_string };
    }
  } catch {
    // args not valid JSON (yet)
  }
  return null;
}

function pretty(json: string): string {
  try {
    return JSON.stringify(JSON.parse(json), null, 2);
  } catch {
    return json;
  }
}

export function ToolCard({ item }: { item: ToolItem }) {
  const diff = editDiff(item.name, item.args);
  return (
    <div className={`tool tool--${item.status}`}>
      <div className="tool__head">
        <span className="tool__name">{item.name}</span>
        {item.readOnly && <span className="tag">read-only</span>}
        <span className={`tool__status tool__status--${item.status}`}>{STATUS_LABEL[item.status]}</span>
      </div>

      {diff ? (
        <DiffView original={diff.original} modified={diff.modified} maxHeight={220} />
      ) : (
        item.args && <CodeViewer value={pretty(item.args)} language="json" maxHeight={160} />
      )}

      {item.output && (
        <div className="tool__out">
          <CodeViewer value={item.output} maxHeight={240} />
          {item.truncated && <div className="tool__note">output truncated</div>}
        </div>
      )}

      {item.error && <div className="tool__err">{item.error}</div>}
    </div>
  );
}
