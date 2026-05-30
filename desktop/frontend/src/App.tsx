import { useState } from "react";
import { useController } from "./lib/useController";
import { Transcript } from "./components/Transcript";
import { Composer } from "./components/Composer";
import { ApprovalModal } from "./components/ApprovalModal";
import { ContextGauge } from "./components/ContextGauge";

export default function App() {
  const { state, send, cancel, approve, setPlan, newSession } = useController();
  const [plan, setPlanLocal] = useState(false);

  const togglePlan = () => {
    const next = !plan;
    setPlanLocal(next);
    setPlan(next);
  };

  return (
    <div className="app">
      <header className="topbar">
        <div className="topbar__brand">Reasonix</div>
        <div className="topbar__model" title="active model">
          {state.meta?.label ?? "…"}
        </div>
        <div className="topbar__spacer" />
        <ContextGauge used={state.context.used} window={state.context.window} />
        <button
          className={`chip ${plan ? "chip--on" : ""}`}
          onClick={togglePlan}
          title="Plan mode — refuse all writes"
        >
          plan
        </button>
        <button className="chip" onClick={newSession} title="Start a new session">
          new
        </button>
      </header>

      {state.meta?.startupErr && (
        <div className="banner banner--error">startup error: {state.meta.startupErr}</div>
      )}

      <main className="main">
        <Transcript items={state.items} />
      </main>

      <footer className="footer">
        <Composer running={state.running} onSend={send} onCancel={cancel} />
      </footer>

      {state.approval && (
        <ApprovalModal
          approval={state.approval}
          onAnswer={(allow, session) => approve(state.approval!.id, allow, session)}
        />
      )}
    </div>
  );
}
