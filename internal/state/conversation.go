package state

import (
	"fmt"
	"sync"

	"expensemate-tgbot/internal/types"
)

// BudgetPendingData stores data for budget set flow
type BudgetPendingData struct {
	Col string // "K" for group, "Q" for category
	Row int    // Sheet row index
}

// VoicePendingData stores data for voice clarification flow
type VoicePendingData struct {
	OriginalText string // The original transcribed text
	ParsedName   string
	ParsedAmount uint64
	ParsedGroup  string
	ParsedCat    string
	ParsedDate   string
	ParsedNote   string
}

// Manager handles conversation state for users
type Manager struct {
	states            map[int64]string
	voicePendingData  map[int64]*VoicePendingData
	budgetPendingData map[int64]*BudgetPendingData
	mu                sync.RWMutex
}

// NewManager creates a new conversation state manager
func NewManager() *Manager {
	return &Manager{
		states:            make(map[int64]string),
		voicePendingData:  make(map[int64]*VoicePendingData),
		budgetPendingData: make(map[int64]*BudgetPendingData),
	}
}

// State constants for conversation flows
const (
	StateExpensesAdd             = "expenses:add"
	StateExpensesVoiceClarify    = "expenses:voice_clarify"
	StateGSheetsConfig           = "gsheets:configure"
	StateGSheetsUpdateActivePage = "gsheets:update_current_page"
	StateBudgetSetGroupPrefix    = "budget:set_group:"
	StateBudgetSetCategoryPrefix = "budget:set_category:"
)

// BuildState creates a state string from command and action
func BuildState(command types.Command, action string) string {
	return fmt.Sprintf("%s:%s", command, action)
}

// Start starts a conversation for a chat
func (m *Manager) Start(chatID int64, state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[chatID] = state
}

// Get returns the current conversation state for a chat
func (m *Manager) Get(chatID int64) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.states[chatID]
}

// End ends the conversation for a chat
func (m *Manager) End(chatID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, chatID)
}

// IsInConversation checks if a chat is in a conversation
func (m *Manager) IsInConversation(chatID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.states[chatID]
	return exists
}

// SetVoicePendingData stores pending voice expense data for clarification
func (m *Manager) SetVoicePendingData(chatID int64, data *VoicePendingData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.voicePendingData[chatID] = data
}

// GetVoicePendingData retrieves pending voice expense data
func (m *Manager) GetVoicePendingData(chatID int64) *VoicePendingData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.voicePendingData[chatID]
}

// ClearVoicePendingData clears pending voice expense data
func (m *Manager) ClearVoicePendingData(chatID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.voicePendingData, chatID)
}

// SetBudgetPendingData stores pending budget data
func (m *Manager) SetBudgetPendingData(chatID int64, data *BudgetPendingData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.budgetPendingData[chatID] = data
}

// GetBudgetPendingData retrieves pending budget data
func (m *Manager) GetBudgetPendingData(chatID int64) *BudgetPendingData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.budgetPendingData[chatID]
}

// ClearBudgetPendingData clears pending budget data
func (m *Manager) ClearBudgetPendingData(chatID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.budgetPendingData, chatID)
}
