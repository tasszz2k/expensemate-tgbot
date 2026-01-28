package state

import (
	"fmt"
	"sync"

	"expensemate-tgbot/internal/types"
)

// Manager handles conversation state for users
type Manager struct {
	states map[int64]string
	mu     sync.RWMutex
}

// NewManager creates a new conversation state manager
func NewManager() *Manager {
	return &Manager{
		states: make(map[int64]string),
	}
}

// State constants for conversation flows
const (
	StateExpensesAdd             = "expenses:add"
	StateGSheetsConfig           = "gsheets:configure"
	StateGSheetsUpdateActivePage = "gsheets:update_current_page"
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
