package input

import (
	"github.com/veandco/go-sdl2/sdl"
)

type keyMapVal struct {
	pressed bool
	changed bool
}

type Keyboard struct {
	keymap map[sdl.Keycode]*keyMapVal
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		keymap: make(map[sdl.Keycode]*keyMapVal),
	}
}

func (k *Keyboard) KeyJustPressed(key sdl.Keycode) bool {
	if val, ok := k.keymap[key]; ok {
		return val.pressed && val.changed
	}
	return false
}

func (k *Keyboard) KeyHeldDown(key sdl.Keycode) bool {
	if val, ok := k.keymap[key]; ok {
		return val.pressed
	}
	return false
}

func (k *Keyboard) KeyReleased(key sdl.Keycode) bool {
	// Can only ever be true if key already exists in keymap
	if val, ok := k.keymap[key]; ok {
		return (! val.pressed) && val.changed
	}
	return false
}

func (k *Keyboard) OnEvent(t *sdl.KeyboardEvent) {
	keyCode := t.Keysym.Sym

	if val, ok := k.keymap[keyCode]; ok {
		if t.State == sdl.PRESSED {
			val.changed = ! val.pressed
			val.pressed = true
		} else if t.State == sdl.RELEASED {
			val.changed = val.pressed
			val.pressed = false
		}
	} else if t.State == sdl.PRESSED {
		// The key code map key does not yet exist
		// so we initially create an entry in the keymap
		k.keymap[keyCode] = &keyMapVal{pressed: true, changed: true}
	}
}

func (k *Keyboard) ResetChangedStates() {
	for _, val := range k.keymap {
		val.changed = false
	}
}