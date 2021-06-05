package gui

import (
	. "github.com/jesseduffield/lazygit/pkg/gui/types"
)

func (gui *Gui) getSelectedSuggestionValue() string {
	selectedSuggestion := gui.getSelectedSuggestion()

	if selectedSuggestion != nil {
		return selectedSuggestion.Value
	}

	return ""
}

func (gui *Gui) getSelectedSuggestion() *Suggestion {
	selectedLine := gui.State.Panels.Suggestions.SelectedLineIdx
	if selectedLine == -1 {
		return nil
	}

	return gui.State.Suggestions[selectedLine]
}

func (gui *Gui) setSuggestions(suggestions []*Suggestion) {
	gui.State.Suggestions = suggestions
	gui.State.Panels.Suggestions.SelectedLineIdx = 0
	_ = gui.resetOrigin(gui.Views.Suggestions)
	_ = gui.State.Contexts.Suggestions.HandleRender()
}
