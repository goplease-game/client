package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ConvertKeyName translates internal Ebiten key string representations
// into clean, human-readable UI text labels.
func ConvertKeyName(k ebiten.Key) string {
	str := k.String()

	// Handle numeric keys (Digit0 - Digit9)
	if len(str) == 6 && str[:5] == "Digit" {
		return str[5:]
	}

	// Handle Numpad digits (Numpad0 - Numpad9)
	if len(str) == 7 && str[:6] == "Numpad" {
		return "Numpad " + str[6:]
	}

	// Handle specific modifier and control keys mapping
	switch str {
	case "AltLeft":
		return "Left Alt"
	case "AltRight":
		return "Right Alt"
	case "ControlLeft":
		return "Left Ctrl"
	case "ControlRight":
		return "Right Ctrl"
	case "ShiftLeft":
		return "Left Shift"
	case "ShiftRight":
		return "Right Shift"
	case "MetaLeft":
		return "Left Win/Cmd"
	case "MetaRight":
		return "Right Win/Cmd"
	case "Space":
		return "Space"
	case "Enter":
		return "Enter"
	case "Escape":
		return "Esc"
	case "ArrowUp":
		return "Up"
	case "ArrowDown":
		return "Down"
	case "ArrowLeft":
		return "Left"
	case "ArrowRight":
		return "Right"
	}

	return str
}

// KeyName checks for unassigned values (-1) and routes
// valid keys through the human-readable formatter.
func KeyName(k *ebiten.Key) string {
	if k == nil || *k < 0 {
		return "[None]"
	}
	return ConvertKeyName(*k)
}

// KeyJustPressed reports whether the bound key was pressed on this frame.
// Safe to call with an unbound (nil) key — always returns false in that case.
func KeyJustPressed(key *ebiten.Key) bool {
	if key == nil {
		return false
	}

	return inpututil.IsKeyJustPressed(*key)
}

// KeyPressed reports whether the bound key is currently held down.
// Safe to call with an unbound (nil) key — always returns false in that case.
func KeyPressed(key *ebiten.Key) bool {
	if key == nil {
		return false
	}

	return ebiten.IsKeyPressed(*key)
}
