package bots

import (
	"context"
)

func (e *Expensemate) updateConversationState(
	_ context.Context,
	chatID int64,
	state string,
) {
	e.csMux.Lock()
	defer e.csMux.Unlock()
	e.conversationStates[chatID] = state
	return
}

func (e *Expensemate) getConversationState(
	_ context.Context,
	chatID int64,
) string {
	e.csMux.RLock()
	defer e.csMux.RUnlock()
	if state, exists := e.conversationStates[chatID]; exists {
		return state
	}
	return ""
}

func (e *Expensemate) removeConversationState(
	_ context.Context,
	chatID int64,
) {
	e.csMux.Lock()
	defer e.csMux.Unlock()
	delete(e.conversationStates, chatID)
	return
}
