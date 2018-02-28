package keyboard

// #include "keyboard.h"
import "C"

func KeyEvent(code byte, up bool) int {
	flag := 0
	if up {
		flag = 1
	}
	C.key_event(C.uchar(code), C.short(flag))
	return flag
}
