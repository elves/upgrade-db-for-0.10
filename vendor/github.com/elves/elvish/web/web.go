// Package web is the entry point for the backend of the web interface of
// Elvish.
package web

//go:generate ./embed-html

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type Web struct {
	ev   *eval.Evaler
	port int
}

type ExecuteResponse struct {
	OutBytes  string
	OutValues []eval.Value
	ErrBytes  string
	Err       string
}

func NewWeb(ev *eval.Evaler, port int) *Web {
	return &Web{ev, port}
}

func (web *Web) Run(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "arguments to -web are not supported yet")
		return 2
	}

	http.HandleFunc("/", web.handleMainPage)
	http.HandleFunc("/execute", web.handleExecute)
	addr := fmt.Sprintf("localhost:%d", web.port)
	log.Println("going to listen", addr)
	err := http.ListenAndServe(addr, nil)

	log.Println(err)
	return 0
}

func (web *Web) handleMainPage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(mainPageHTML))
	if err != nil {
		log.Println("cannot write response:", err)
	}
}

func (web *Web) handleExecute(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("cannot read request body:", err)
		return
	}
	text := string(bytes)

	outBytes, outValues, errBytes, err := evalAndCollect(web.ev, "<web>", text)
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	responseBody, err := json.Marshal(
		&ExecuteResponse{string(outBytes), outValues, string(errBytes), errText})
	if err != nil {
		log.Println("cannot marshal response body:", err)
	}

	_, err = w.Write(responseBody)
	if err != nil {
		log.Println("cannot write response:", err)
	}
}

const (
	outFileBufferSize = 1024
	outChanBufferSize = 32
)

// evalAndCollect evaluates a piece of code with null stdin, and stdout and
// stderr connected to pipes (value part of stderr being a blackhole), and
// return the results collected on stdout and stderr, and the possible error
// that occurred.
func evalAndCollect(ev *eval.Evaler, name, text string) (
	outBytes []byte, outValues []eval.Value, errBytes []byte, err error) {

	node, err := parse.Parse(name, text)
	if err != nil {
		return
	}
	op, err := ev.Compile(node, name, text)
	if err != nil {
		return
	}

	outFile, chanOutBytes := makeBytesWriterAndCollect()
	outChan, chanOutValues := makeValuesWriterAndCollect()
	errFile, chanErrBytes := makeBytesWriterAndCollect()

	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{File: outFile, Chan: outChan},
		{File: errFile, Chan: eval.BlackholeChan},
	}
	err = ev.EvalWithPorts(ports, op, name, text)

	outFile.Close()
	close(outChan)
	errFile.Close()
	return <-chanOutBytes, <-chanOutValues, <-chanErrBytes, err
}

// makeBytesWriterAndCollect makes an in-memory file that can be written to, and
// the written bytes will be collected in a byte slice that will be put on a
// channel as soon as the writer is closed.
func makeBytesWriterAndCollect() (*os.File, <-chan []byte) {
	r, w, err := os.Pipe()
	// os.Pipe returns error only on resource exhaustion.
	if err != nil {
		panic(err)
	}
	chanCollected := make(chan []byte)

	go func() {
		var (
			collected []byte
			buf       [outFileBufferSize]byte
		)
		for {
			n, err := r.Read(buf[:])
			collected = append(collected, buf[:n]...)
			if err != nil {
				if err != io.EOF {
					log.Println("error when reading output pipe:", err)
				}
				break
			}
		}
		r.Close()
		chanCollected <- collected
	}()

	return w, chanCollected
}

// makeValuesWriterAndCollect makes a Value channel for writing, and the written
// values will be collected in a Value slice that will be put on a channel as
// soon as the writer is closed.
func makeValuesWriterAndCollect() (chan eval.Value, <-chan []eval.Value) {
	chanValues := make(chan eval.Value, outChanBufferSize)
	chanCollected := make(chan []eval.Value)

	go func() {
		var collected []eval.Value
		for {
			for v := range chanValues {
				collected = append(collected, v)
			}
			chanCollected <- collected
		}
	}()

	return chanValues, chanCollected
}
