package lib

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	keyNUL tea.KeyType = 0
	keySOH tea.KeyType = 1
	keySTX tea.KeyType = 2
	keyETX tea.KeyType = 3
	keyEOT tea.KeyType = 4
	keyENQ tea.KeyType = 5
	keyACK tea.KeyType = 6
	keyBEL tea.KeyType = 7
	keyBS  tea.KeyType = 8
	keyHT  tea.KeyType = 9
	keyLF  tea.KeyType = 10
	keyVT  tea.KeyType = 11
	keyFF  tea.KeyType = 12
	keyCR  tea.KeyType = 13
	keySO  tea.KeyType = 14
	keySI  tea.KeyType = 15
	keyDLE tea.KeyType = 16
	keyDC1 tea.KeyType = 17
	keyDC2 tea.KeyType = 18
	keyDC3 tea.KeyType = 19
	keyDC4 tea.KeyType = 20
	keyNAK tea.KeyType = 21
	keySYN tea.KeyType = 22
	keyETB tea.KeyType = 23
	keyCAN tea.KeyType = 24
	keyEM  tea.KeyType = 25
	keySUB tea.KeyType = 26
	keyESC tea.KeyType = 27
	keyFS  tea.KeyType = 28
	keyGS  tea.KeyType = 29
	keyRS  tea.KeyType = 30
	keyUS  tea.KeyType = 31
	keyDEL tea.KeyType = 127
)

var keyNames = map[string]tea.KeyType{

	"ctrl+@":    keyNUL,
	"ctrl+a":    keySOH,
	"ctrl+b":    keySTX,
	"ctrl+c":    keyETX,
	"ctrl+d":    keyEOT,
	"ctrl+e":    keyENQ,
	"ctrl+f":    keyACK,
	"ctrl+g":    keyBEL,
	"ctrl+h":    keyBS,
	"tab":       keyHT,
	"ctrl+j":    keyLF,
	"ctrl+k":    keyVT,
	"ctrl+l":    keyFF,
	"enter":     keyCR,
	"ctrl+n":    keySO,
	"ctrl+o":    keySI,
	"ctrl+p":    keyDLE,
	"ctrl+q":    keyDC1,
	"ctrl+r":    keyDC2,
	"ctrl+s":    keyDC3,
	"ctrl+t":    keyDC4,
	"ctrl+u":    keyNAK,
	"ctrl+v":    keySYN,
	"ctrl+w":    keyETB,
	"ctrl+x":    keyCAN,
	"ctrl+y":    keyEM,
	"ctrl+z":    keySUB,
	"esc":       keyESC,
	"ctrl+\\":   keyFS,
	"ctrl+]":    keyGS,
	"ctrl+^":    keyRS,
	"ctrl+_":    keyUS,
	"backspace": keyDEL,

	"runes":            tea.KeyRunes,
	"up":               tea.KeyUp,
	"down":             tea.KeyDown,
	"right":            tea.KeyRight,
	" ":                tea.KeySpace,
	"left":             tea.KeyLeft,
	"shift+tab":        tea.KeyShiftTab,
	"home":             tea.KeyHome,
	"end":              tea.KeyEnd,
	"pgup":             tea.KeyPgUp,
	"pgdown":           tea.KeyPgDown,
	"delete":           tea.KeyDelete,
	"ctrl+up":          tea.KeyCtrlUp,
	"ctrl+down":        tea.KeyCtrlDown,
	"ctrl+right":       tea.KeyCtrlRight,
	"ctrl+left":        tea.KeyCtrlLeft,
	"shift+up":         tea.KeyShiftUp,
	"shift+down":       tea.KeyShiftDown,
	"shift+right":      tea.KeyShiftRight,
	"shift+left":       tea.KeyShiftLeft,
	"ctrl+shift+up":    tea.KeyCtrlShiftUp,
	"ctrl+shift+down":  tea.KeyCtrlShiftDown,
	"ctrl+shift+left":  tea.KeyCtrlShiftLeft,
	"ctrl+shift+right": tea.KeyCtrlShiftRight,
	"f1":               tea.KeyF1,
	"f2":               tea.KeyF2,
	"f3":               tea.KeyF3,
	"f4":               tea.KeyF4,
	"f5":               tea.KeyF5,
	"f6":               tea.KeyF6,
	"f7":               tea.KeyF7,
	"f8":               tea.KeyF8,
	"f9":               tea.KeyF9,
	"f10":              tea.KeyF10,
	"f11":              tea.KeyF11,
	"f12":              tea.KeyF12,
	"f13":              tea.KeyF13,
	"f14":              tea.KeyF14,
	"f15":              tea.KeyF15,
	"f16":              tea.KeyF16,
	"f17":              tea.KeyF17,
	"f18":              tea.KeyF18,
	"f19":              tea.KeyF19,
	"f20":              tea.KeyF20,
}

func getKey(input string) (tea.Key, error) {
	var key tea.Key
	if strings.HasPrefix(input, "alt+") {
		key.Alt = true
		input = input[4:]
	}
	if len(input) == 0 {
		return key, errors.New("no string to create a key from")
	}

	kt, ok := keyNames[input]
	if !ok {
		key.Runes = []rune(input)
		key.Type = tea.KeyRunes
		return key, nil
	}
	key.Type = kt
	if key.Type == tea.KeySpace {
		key.Runes = []rune{' '}
	}
	return key, nil
}
