package input

import (
	"github.com/veandco/go-sdl2/sdl"
)

type keyMapVal struct {
	pressed bool
	changed bool
}

type Keyboard struct {
	kmap map[sdl.Keycode]*keyMapVal
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		kmap: make(map[sdl.Keycode]*keyMapVal),
	}
}

func (k *Keyboard) KeyJustPressed(key sdl.Keycode) bool {
	if val, ok := k.kmap[key]; ok {
		return val.pressed && val.changed
	}
	return false
}

func (k *Keyboard) KeyHeldDown(key sdl.Keycode) bool {
	if val, ok := k.kmap[key]; ok {
		return val.pressed
	}
	return false
}

func (k *Keyboard) KeyReleased(key sdl.Keycode) bool {
	// Can only ever be true if key already exists in keymap
	if val, ok := k.kmap[key]; ok {
		return (! val.pressed) && val.changed
	}
	return false
}

func (k *Keyboard) OnEvent(t *sdl.KeyboardEvent) {
	keyCode := t.Keysym.Sym

	if val, ok := k.kmap[keyCode]; ok {
		if t.State == sdl.PRESSED {
			val.changed = ! val.pressed
			val.pressed = true
		} else if t.State == sdl.RELEASED {
			val.changed = val.pressed
			val.pressed = false
		}
	} else if t.State == sdl.PRESSED {
		k.kmap[keyCode] = &keyMapVal{pressed: true, changed: true}
	}
}

func (k *Keyboard) ResetChangedStates() {
	for _, val := range k.kmap {
		val.changed = false
	}
}