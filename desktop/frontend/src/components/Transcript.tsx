import { useEffect, useRef } from "react";
import type { Item } from "../lib/useController";
import { AssistantMessage, UserMessage } from "./Message";
import { ToolCard } from "./ToolCard";

export function Transcript({ items }: { items: Item[] }) {
  const endRef = useRef<HTMLDivElement>(null);

  // Keep the newest content in view as the turn streams.
  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" });
  }, [items]);

  return (
    <div className="transcript">
      {items.length === 0 && (
        <div className="empty">
          <div className="empty__title">Reasonix</div>
          <div className="empty__hint">
            Ask anything, or describe a task. Reasoning, tool calls, and approvals appear inline.
          </div>
        </div>
      )}

      {items.map((it) => {
        switch (it.kind) {
          case "user":
            return <UserMessage key={it.id} text={it.text} />;
          case "assistant":
            return <AssistantMessage key={it.id} item={it} />;
          case "tool":
            return <ToolCard key={it.id} item={it} />;
          case "phase":
            return (
              <div key={it.id} className="phase">
                {it.text}
              </div>
            );
          case "notice":
            return (
              <div key={it.id} className={`notice notice--${it.level}`}>
                {it.text}
              </div>
            );
        }
      })}

      <div ref={endRef} />
    </div>
  );
}
