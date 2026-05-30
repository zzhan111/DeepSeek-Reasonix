import { useState } from "react";
import type { KeyboardEvent } from "react";

export function Composer({
  running,
  onSend,
  onCancel,
}: {
  running: boolean;
  onSend: (text: string) => void;
  onCancel: () => void;
}) {
  const [text, setText] = useState("");

  const submit = () => {
    const t = text.trim();
    if (!t) return;
    onSend(t);
    setText("");
  };

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    // Enter sends; Shift+Enter inserts a newline. isComposing guards IME input
    // (e.g. pinyin) so confirming a candidate doesn't fire a send.
    if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
      e.preventDefault();
      submit();
    }
  };

  return (
    <div className="composer">
      <textarea
        className="composer__input"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder="Message Reasonix…  (Enter to send · Shift+Enter for newline)"
        rows={1}
      />
      {running ? (
        <button className="btn btn--stop" onClick={onCancel}>
          Stop
        </button>
      ) : (
        <button className="btn btn--send" onClick={submit} disabled={!text.trim()}>
          Send
        </button>
      )}
    </div>
  );
}
