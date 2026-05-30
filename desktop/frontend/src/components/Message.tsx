import { useState } from "react";
import { Markdown } from "./Markdown";
import type { Item } from "../lib/useController";

type AssistantItem = Extract<Item, { kind: "assistant" }>;

export function UserMessage({ text }: { text: string }) {
  return (
    <div className="msg msg--user">
      <div className="msg__body">{text}</div>
    </div>
  );
}

export function AssistantMessage({ item }: { item: AssistantItem }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="msg msg--assistant">
      {item.reasoning && (
        <div className="reasoning">
          <button className="reasoning__toggle" onClick={() => setOpen((v) => !v)}>
            {open ? "▾" : "▸"} thinking
          </button>
          {open && <div className="reasoning__body">{item.reasoning}</div>}
        </div>
      )}
      <div className="msg__body">
        <Markdown text={item.text} />
        {item.streaming && <span className="cursor" />}
      </div>
    </div>
  );
}
