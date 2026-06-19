package etch

func CheckInvalidHeaderChar(val string) bool {
	for i, l := 0, len(val); i < l; i++ {
		b := val[i]
		if isCTL(b) && !isLWS(b) {
			return true
		}
	}
	return false
}

func isLWS(b byte) bool { return b == ' ' || b == '\t' }

func isCTL(b byte) bool {
	return b < ' ' || b == 0x7f
}
