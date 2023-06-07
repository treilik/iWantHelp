package lib

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type input struct {
	model textinput.Model
}

func (i input) Init() tea.Cmd { return nil }
func (i input) View() string  { return i.model.View() }
func (i input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := i.model.Update(msg)
	i.model = m
	return i, cmd
}
func (w *Walder) activateInput(submitFunc func(w *Walder, input string) error) error {
	model, ok := w.ui.ModelMap[inputAddr]
	if !ok || model == nil {
		return fmt.Errorf("no input found")
	}
	inputModel := model.(input)
	inputModel.model.Focus()
	w.ui.ModelMap[inputAddr] = inputModel

	w.inputFunc = submitFunc
	return nil
}
func (w *Walder) deactivateInput() (func(w *Walder) error, error) {
	if w.inputFunc == nil {
		return nil, fmt.Errorf("no input function set")
	}
	model, ok := w.ui.ModelMap[inputAddr]
	if !ok || model == nil {
		return nil, fmt.Errorf("no input found")
	}
	inputModel := model.(input)
	input := inputModel.model.Value()
	inputModel.model.Reset()
	inputModel.model.Blur()
	w.ui.ModelMap[inputAddr] = inputModel

	inputFunc := w.inputFunc
	w.inputFunc = nil
	return func(w *Walder) error {
		return inputFunc(w, input)
	}, nil
}
