package primo

const (
	ptpHeader          = 0xff
	dataBlockHeader    = 0x55
	closingBlockHeader = 0xaa
)

type TapePlayer struct {
	tape      []byte
	bytePos   int
	blockSize uint16
}

func NewTapePlayer() *TapePlayer {
	return &TapePlayer{}
}

func (t *TapePlayer) ChangeTape(tape []byte) {
	t.tape = tape
	t.Reset()
}

func (t *TapePlayer) Reset() {
	t.bytePos = 0
	t.blockSize = 0
}

func (t *TapePlayer) readBlockHeader() bool {
	// we can just skip the PTP header
	if t.tape[t.bytePos] == ptpHeader {
		if t.bytePos+3 >= len(t.tape) {
			return false
		}
		t.bytePos += 3
	}

	// the next byte should either be a data or a closing block header
	if t.tape[t.bytePos] != dataBlockHeader && t.tape[t.bytePos] != closingBlockHeader {
		return false
	}

	// read the size of the next block
	if t.bytePos+3 >= len(t.tape) {
		return false
	}
	t.blockSize = uint16(t.tape[t.bytePos+1]) | (uint16(t.tape[t.bytePos+2]) << 8)
	t.bytePos += 3

	return true
}

func (t *TapePlayer) NextByte() byte {
	if len(t.tape) == 0 {
		return 0
	}

	// if we got to the end just restart the tape
	if t.bytePos == len(t.tape) {
		t.Reset()
	}

	if t.blockSize == 0 && !t.readBlockHeader() {
		// invalid PTP file
		return 0
	}

	t.blockSize--
	t.bytePos++
	return t.tape[t.bytePos-1]
}
