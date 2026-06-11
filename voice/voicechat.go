package voicechat

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/AnthonyL103/GOMCP/chat"
)

const (
	defaultSilenceDebounceMs = 900
	defaultHardFinalizeMs    = 3000
)

type JobType string

const (
	JobNone           JobType = "none"
	JobLLMResponse    JobType = "llm_response"
	JobToolGeneration JobType = "tool_generation"
)

type InterruptPolicy struct {
	ImmediateCancelForLLM bool
	SafePauseForToolGen   bool
	SilenceMs             int
	HardFinalizeMs        int
	AutoResume            bool
}

func DefaultInterruptPolicy() InterruptPolicy {
	return InterruptPolicy{
		ImmediateCancelForLLM: true,
		SafePauseForToolGen:   true,
		SilenceMs:             defaultSilenceDebounceMs,
		HardFinalizeMs:        defaultHardFinalizeMs,
		AutoResume:            false,
	}
}

type VoiceSessionSnapshot struct {
	Active         bool
	Transcript     string
	LastSpeechAt   time.Time
	StartedAt      time.Time
	DebounceMs     int
	HardFinalizeMs int
	PauseRequested bool
	PauseReason    string
}

type VoiceSessionState struct {
	mu sync.RWMutex

	active         bool
	transcript     string
	lastSpeechAt   time.Time
	startedAt      time.Time
	debounceMs     int
	hardFinalizeMs int

	pauseRequested bool
	pauseReason    string
}

func NewVoiceSessionState(policy InterruptPolicy) *VoiceSessionState {
	silence := policy.SilenceMs
	if silence <= 0 {
		silence = defaultSilenceDebounceMs
	}
	hard := policy.HardFinalizeMs
	if hard <= 0 {
		hard = defaultHardFinalizeMs
	}

	return &VoiceSessionState{
		debounceMs:     silence,
		hardFinalizeMs: hard,
	}
}

func (s *VoiceSessionState) StartSpeech(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	if !s.active {
		s.active = true
		s.startedAt = now
		s.transcript = ""
	}
	s.lastSpeechAt = now
}

func (s *VoiceSessionState) MarkSpeech(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	if !s.active {
		s.active = true
		s.startedAt = now
	}
	s.lastSpeechAt = now
}

func (s *VoiceSessionState) AppendTranscript(chunk string, now time.Time) {
	chunk = strings.TrimSpace(chunk)
	if chunk == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	if !s.active {
		s.active = true
		s.startedAt = now
	}

	if s.transcript == "" {
		s.transcript = chunk
	} else {
		s.transcript = strings.TrimSpace(s.transcript + " " + chunk)
	}

	s.lastSpeechAt = now
}

func (s *VoiceSessionState) RequestPause(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pauseRequested = true
	s.pauseReason = strings.TrimSpace(reason)
}

func (s *VoiceSessionState) ClearPauseRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pauseRequested = false
	s.pauseReason = ""
}

func (s *VoiceSessionState) ShouldFinalize(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.active || s.lastSpeechAt.IsZero() {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}

	return now.Sub(s.lastSpeechAt) >= time.Duration(s.debounceMs)*time.Millisecond
}

func (s *VoiceSessionState) ShouldForceFinalize(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.active || s.startedAt.IsZero() {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}

	return now.Sub(s.startedAt) >= time.Duration(s.hardFinalizeMs)*time.Millisecond
}

func (s *VoiceSessionState) FinalizeUtterance() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	final := strings.TrimSpace(s.transcript)
	s.active = false
	s.transcript = ""
	s.lastSpeechAt = time.Time{}
	s.startedAt = time.Time{}
	s.pauseRequested = false
	s.pauseReason = ""

	return final
}

func (s *VoiceSessionState) Snapshot() VoiceSessionSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return VoiceSessionSnapshot{
		Active:         s.active,
		Transcript:     s.transcript,
		LastSpeechAt:   s.lastSpeechAt,
		StartedAt:      s.startedAt,
		DebounceMs:     s.debounceMs,
		HardFinalizeMs: s.hardFinalizeMs,
		PauseRequested: s.pauseRequested,
		PauseReason:    s.pauseReason,
	}
}

func (s *VoiceSessionState) SetDebounceMs(ms int) {
	if ms <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debounceMs = ms
}

func (s *VoiceSessionState) SetHardFinalizeMs(ms int) {
	if ms <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hardFinalizeMs = ms
}

type RuntimeExecutionSnapshot struct {
	CurrentJobID string
	CurrentType  JobType
	CurrentStage string
	Paused       bool
	PauseReason  string
	PausedAt     time.Time
	Cancellable  bool
	LastMessage  chat.Message
	LastToolCall chat.ToolCall
}

type RuntimeExecutionState struct {
	mu sync.RWMutex

	currentJobID string
	currentType  JobType
	currentStage string
	paused       bool
	pauseReason  string
	pausedAt     time.Time
	lastMessage  chat.Message
	lastToolCall chat.ToolCall

	cancelFn context.CancelFunc
}

func NewRuntimeExecutionState() *RuntimeExecutionState {
	return &RuntimeExecutionState{currentType: JobNone}
}

func (s *RuntimeExecutionState) BeginJob(jobID string, jobType JobType, stage string, cancelFn context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentJobID = strings.TrimSpace(jobID)
	s.currentType = jobType
	s.currentStage = strings.TrimSpace(stage)
	s.paused = false
	s.pauseReason = ""
	s.pausedAt = time.Time{}
	s.cancelFn = cancelFn
}

func (s *RuntimeExecutionState) UpdateStage(stage string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentStage = strings.TrimSpace(stage)
}

func (s *RuntimeExecutionState) RequestPause(reason string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	s.paused = true
	s.pauseReason = strings.TrimSpace(reason)
	s.pausedAt = now
}

func (s *RuntimeExecutionState) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = false
	s.pauseReason = ""
	s.pausedAt = time.Time{}
}

func (s *RuntimeExecutionState) CompleteJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobID = strings.TrimSpace(jobID)
	if jobID != "" && s.currentJobID != jobID {
		return
	}

	s.currentJobID = ""
	s.currentType = JobNone
	s.currentStage = ""
	s.paused = false
	s.pauseReason = ""
	s.pausedAt = time.Time{}
	s.cancelFn = nil
}

func (s *RuntimeExecutionState) InterruptForVoice(policy InterruptPolicy, reason string, now time.Time) (canceled bool, paused bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "voice interrupt"
	}

	s.paused = true
	s.pauseReason = reason
	s.pausedAt = now

	if s.currentType == JobLLMResponse && policy.ImmediateCancelForLLM && s.cancelFn != nil {
		s.cancelFn()
		canceled = true
	}

	if s.currentType == JobToolGeneration && policy.SafePauseForToolGen {
		paused = true
	}

	if s.currentType == JobNone {
		paused = true
	}

	return canceled, paused
}

func (s *RuntimeExecutionState) IsPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paused
}

func (s *RuntimeExecutionState) Snapshot() RuntimeExecutionSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return RuntimeExecutionSnapshot{
		CurrentJobID: s.currentJobID,
		CurrentType:  s.currentType,
		CurrentStage: s.currentStage,
		Paused:       s.paused,
		PauseReason:  s.pauseReason,
		PausedAt:     s.pausedAt,
		Cancellable:  s.cancelFn != nil,
	}
}
