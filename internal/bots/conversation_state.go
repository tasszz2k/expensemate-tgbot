package bots

import (
	"context"
)

func (e *Expensemate) updateConversationState(
	ctx context.Context,
	chatID int64,
	state string,
) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.conversationStates[chatID] = state
	return
}

func (e *Expensemate) getConversationState(
	ctx context.Context,
	chatID int64,
) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if state, exists := e.conversationStates[chatID]; exists {
		return state
	}
	return ""
}

func (e *Expensemate) removeConversationState(
	ctx context.Context,
	chatID int64,
) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.conversationStates, chatID)
	return
}
