package eval

import "os"

// Port conveys data stream. It always consists of a byte band and a channel band.
type Port struct {
	File      *os.File
	Chan      chan Value
	CloseFile bool
	CloseChan bool
}

// Fork returns a copy of a Port with the Close* flags unset.
func (p *Port) Fork() *Port {
	return &Port{p.File, p.Chan, false, false}
}

// Close closes a Port.
func (p *Port) Close() {
	if p == nil {
		return
	}
	if p.CloseFile {
		p.File.Close()
	}
	if p.CloseChan {
		// Logger.Printf("closing channel %v", p.Chan)
		close(p.Chan)
	}
}

// ClosePorts closes a list of Ports.
func ClosePorts(ports []*Port) {
	for _, port := range ports {
		// Logger.Printf("closing port %d", i)
		port.Close()
	}
}

var (
	// ClosedChan is a closed channel, suitable for use as placeholder channel input.
	ClosedChan = make(chan Value)
	// BlackholeChan is channel writes onto which disappear, suitable for use as
	// placeholder channel output.
	BlackholeChan = make(chan Value)
	// DevNull is /dev/null.
	DevNull *os.File
	// DevNullClosedInput is a port made up from DevNull and ClosedChan,
	// suitable as placeholder input port.
	DevNullClosedChan *Port
)

func init() {
	close(ClosedChan)
	go func() {
		for range BlackholeChan {
		}
	}()

	var err error
	DevNull, err = os.Open("/dev/null")
	if err != nil {
		os.Stderr.WriteString("cannot open /dev/null, shell might not function normally")
	}
	DevNullClosedChan = &Port{File: DevNull, Chan: ClosedChan}
}
