import type { WireApproval } from "../lib/types";

export function ApprovalModal({
  approval,
  onAnswer,
}: {
  approval: WireApproval;
  onAnswer: (allow: boolean, session: boolean) => void;
}) {
  return (
    <div className="modal-backdrop">
      <div className="modal">
        <div className="modal__title">Allow this tool call?</div>
        <div className="modal__tool">
          <span className="tool__name">{approval.tool}</span>
        </div>
        {approval.subject && <pre className="modal__subject">{approval.subject}</pre>}
        <div className="modal__actions">
          <button className="btn" onClick={() => onAnswer(false, false)}>
            Deny
          </button>
          <button className="btn" onClick={() => onAnswer(true, false)}>
            Allow once
          </button>
          <button className="btn btn--primary" onClick={() => onAnswer(true, true)}>
            Allow for session
          </button>
        </div>
      </div>
    </div>
  );
}
