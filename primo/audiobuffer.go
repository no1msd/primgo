package primo

const (
	pcmLow  = 0x0
	pcmHigh = 0x4000
)

type AudioBuffer struct {
	buf        []byte
	sampleRate int
}

func NewAudioBuffer(sampleRate int) *AudioBuffer {
	return &AudioBuffer{sampleRate: sampleRate}
}

// Read is the io.Reader implementation for the audio player to read the PCM stream.
func (a *AudioBuffer) Read(p []byte) (int, error) {
	if len(a.buf) == 0 {
		// if the buffer is empty send 5ms of silence
		for i := 0; i < (a.sampleRate/200)*4; i++ {
			a.buf = append(a.buf, 0)
		}
	}

	length := copy(p, a.buf)
	a.buf = a.buf[:0]
	return length, nil
}

func (a *AudioBuffer) AddSample(high bool) {
	smp := pcmLow
	if high {
		smp = pcmHigh
	}

	// audio stream should be 16bit little endian 2 channel stereo PCM
	a.buf = append(a.buf, []byte{byte(smp), byte(smp >> 8), byte(smp), byte(smp >> 8)}...)
}
