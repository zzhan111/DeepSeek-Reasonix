package main

import (
	"context"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"reasonix/internal/agent"
	"reasonix/internal/boot"
	"reasonix/internal/control"
	"reasonix/internal/event"
	"reasonix/internal/memory"
)

// eventChannel is the Wails runtime event name the frontend subscribes to for the
// agent's typed event stream. One channel carries every event kind; the payload's
// `kind` field discriminates — the desktop analogue of the serve transport's SSE
// `data:` frames.
const eventChannel = "agent:event"

// App is the Wails-bound application object: the desktop frontend's command
// surface. Its exported methods (Submit/Cancel/Approve/…) are generated into JS
// bindings and call straight through to one transport-agnostic control.Controller
// — the same controller the chat TUI and the HTTP/SSE server drive, assembled by
// the shared internal/boot. Events flow the other way: the controller emits to an
// eventSink that forwards each one to the webview via runtime.EventsEmit.
type App struct {
	ctx  context.Context
	sink *eventSink
	ctrl *control.Controller

	startupErr string
	label      string
}

// NewApp constructs the bound object. The controller is built later, in startup,
// once the Wails context exists.
func NewApp() *App { return &App{sink: &eventSink{}} }

// startup runs once the webview process is up, before the frontend can issue any
// bound call. It captures the Wails context (needed for EventsEmit), points the
// sink at it, then builds the controller with that sink — so the event bridge is
// live before the first command lands. RequireKey is false so a missing API key
// opens the window in a "set your key" state rather than failing to launch; a
// build error is surfaced through Meta instead of crashing the window.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.sink.ctx = ctx

	ctrl, err := boot.Build(ctx, boot.Options{RequireKey: false, Sink: a.sink})
	if err != nil {
		a.startupErr = err.Error()
		return
	}
	a.ctrl = ctrl
	a.label = ctrl.Label()

	// Desktop is interactive: route "ask" gate decisions to the frontend as
	// approval_request events, answered via Approve.
	ctrl.EnableInteractiveApproval()

	// Land auto-save in a fresh session file (same as a fresh chat/serve start).
	if dir := ctrl.SessionDir(); dir != "" {
		ctrl.SetSessionPath(agent.NewSessionPath(dir, ctrl.Label()))
	}
}

// shutdown snapshots the conversation and stops plugin subprocesses on close.
func (a *App) shutdown(context.Context) {
	if a.ctrl != nil {
		_ = a.ctrl.Snapshot()
		a.ctrl.Close()
	}
}

// --- bound command surface (frontend → controller) ---
// Each method guards on a nil controller so a pre-startup or failed-build call is
// a no-op, never a panic.

// Submit runs raw user input as a turn; slash commands and @-references are
// resolved by the controller. Output arrives asynchronously on eventChannel.
func (a *App) Submit(input string) {
	if a.ctrl != nil {
		a.ctrl.Submit(input)
	}
}

// Cancel aborts the in-flight turn.
func (a *App) Cancel() {
	if a.ctrl != nil {
		a.ctrl.Cancel()
	}
}

// Approve answers a pending approval_request by ID: allow runs the call, session
// also remembers the grant for the rest of the session.
func (a *App) Approve(id string, allow, session bool) {
	if a.ctrl != nil {
		a.ctrl.Approve(id, allow, session)
	}
}

// SetPlanMode toggles read-only plan mode.
func (a *App) SetPlanMode(on bool) {
	if a.ctrl != nil {
		a.ctrl.SetPlanMode(on)
	}
}

// Compact runs one compaction pass on demand.
func (a *App) Compact() error {
	if a.ctrl == nil {
		return nil
	}
	return a.ctrl.Compact(a.ctx)
}

// NewSession snapshots the current conversation and rotates to a fresh one.
func (a *App) NewSession() error {
	if a.ctrl == nil {
		return nil
	}
	return a.ctrl.NewSession()
}

// --- memory panel (frontend ⇄ controller) ---

// MemoryDoc is one loaded doc-memory file for the panel: its path, scope, and
// body. The frontend lists these so the user can see and open each REASONIX.md.
type MemoryDoc struct {
	Path  string `json:"path"`
	Scope string `json:"scope"`
	Body  string `json:"body"`
}

// MemoryFact is one saved auto-memory, surfaced in the panel's "saved" list.
type MemoryFact struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Body        string `json:"body"`
}

// MemoryView is the whole memory panel payload: hierarchical docs plus saved
// facts, with the store path so the UI can offer "reveal in finder".
type MemoryView struct {
	Docs      []MemoryDoc  `json:"docs"`
	Facts     []MemoryFact `json:"facts"`
	StoreDir  string       `json:"storeDir"`
	Available bool         `json:"available"`
}

// Memory returns the loaded memory for the panel: the REASONIX.md hierarchy and
// the saved auto-memories. Read-only; mutations go through Remember / Submit's
// "#" shortcut.
func (a *App) Memory() MemoryView {
	if a.ctrl == nil {
		return MemoryView{}
	}
	set := a.ctrl.Memory()
	if set == nil {
		return MemoryView{}
	}
	view := MemoryView{StoreDir: set.Store.Dir, Available: true}
	for _, d := range set.Docs {
		view.Docs = append(view.Docs, MemoryDoc{Path: d.Path, Scope: string(d.Scope), Body: d.Body})
	}
	for _, f := range set.Store.List() {
		view.Facts = append(view.Facts, MemoryFact{
			Name: f.Name, Description: f.Description, Type: string(f.Type), Body: f.Body,
		})
	}
	return view
}

// Remember quick-adds a one-line note to the project REASONIX.md — the panel's
// explicit "remember" button, equivalent to typing "#<note>". Returns the file
// written so the UI can confirm. It also takes effect on the next turn this
// session without touching the cached prompt prefix.
func (a *App) Remember(note string) (string, error) {
	if a.ctrl == nil {
		return "", nil
	}
	return a.ctrl.QuickAdd(memory.ScopeProject, note)
}

// HistoryMessage is one prior turn, for the frontend to repopulate its transcript
// after a reload.
type HistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// History returns the session's message log.
func (a *App) History() []HistoryMessage {
	if a.ctrl == nil {
		return nil
	}
	msgs := a.ctrl.History()
	out := make([]HistoryMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, HistoryMessage{Role: string(m.Role), Content: m.Content})
	}
	return out
}

// ContextInfo is the prompt-vs-window gauge payload. Both zero means no data yet.
type ContextInfo struct {
	Used   int `json:"used"`
	Window int `json:"window"`
}

// ContextUsage returns the latest context-window gauge numbers.
func (a *App) ContextUsage() ContextInfo {
	if a.ctrl == nil {
		return ContextInfo{}
	}
	used, window := a.ctrl.ContextSnapshot()
	return ContextInfo{Used: used, Window: window}
}

// Meta describes the session for the frontend's header and status line.
type Meta struct {
	Label        string `json:"label"`
	Ready        bool   `json:"ready"`
	StartupErr   string `json:"startupErr,omitempty"`
	EventChannel string `json:"eventChannel"`
	Cwd          string `json:"cwd"`
}

// Meta reports the model label, readiness, any startup error, the working
// directory (for the status line), and the runtime event channel the frontend
// subscribes to.
func (a *App) Meta() Meta {
	cwd, _ := os.Getwd()
	return Meta{
		Label:        a.label,
		Ready:        a.ctrl != nil,
		StartupErr:   a.startupErr,
		EventChannel: eventChannel,
		Cwd:          cwd,
	}
}

// eventSink is the controller's event.Sink in desktop mode: it forwards every
// agent event to the webview as one runtime event, JSON-shaped by toWire. It is a
// type distinct from App so App's bound method set stays the clean command surface
// — Emit must not be exposed to JS. Emit runs on the agent goroutine;
// runtime.EventsEmit is goroutine-safe, and the ctx guard covers the brief window
// before startup assigns it.
type eventSink struct{ ctx context.Context }

func (s *eventSink) Emit(e event.Event) {
	if s.ctx == nil {
		return
	}
	runtime.EventsEmit(s.ctx, eventChannel, toWire(e))
}
