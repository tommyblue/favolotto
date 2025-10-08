package favolotto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tommyblue/favolotto/internal/colors"
)

const (
	stopped int32 = iota
	playing
	paused
)

type Audio struct {
	in        <-chan string
	ctrl      <-chan string
	ledColor  chan<- colors.Color
	storePath string

	cmd         string
	currentFile string
	currentCmd  *exec.Cmd
	volume      int
	status      int32
}

// List of the possible control commands
var (
	Pause  = "pause"
	Resume = "resume"
	Stop   = "stop"
	Volume = "volume"
)

var volumes = []int{20, 60, 100}

func NewAudio(storePath string, in <-chan string, ctrl <-chan string, ledColor chan<- colors.Color) (*Audio, error) {
	cmd := "mpg321"
	_, err := exec.LookPath(cmd)
	if err != nil {
		log.Println("mpg321 not found, trying mpg123")
		cmd = "mpg123"
		_, err = exec.LookPath(cmd)
		if err != nil {
			return nil, fmt.Errorf("neither mpg321 nor mpg123 found")
		}
	}

	return &Audio{
		in:        in,
		ctrl:      ctrl,
		ledColor:  ledColor,
		storePath: storePath,
		cmd:       cmd,
		volume:    100,
	}, nil
}

func (a *Audio) Run(ctx context.Context) {
	stdin := a.setup()
	for {
		select {
		case <-ctx.Done():
			a.cleanup()
			log.Println("Audio context done")
			return
		case fname := <-a.in:
			if fname == "" && a.currentFile != "" {
				a.stop(stdin)
				break
			}

			// if the same file is requested, do nothing
			// TODO: if the file is the same but the audio is paused, resume it
			if a.currentFile == fname {
				log.Println("Same audio file requested")
				// a.ledColor <- "orange"
				// go func() {
				// 	time.Sleep(2 * time.Second)
				// 	a.ledColor <- "black"
				// }()
				break
			}
			log.Println("Audio file requested: ", fname)

			// check if the file exists
			if _, err := os.Stat(fname); os.IsNotExist(err) {
				log.Printf("File %s does not exist", fname)
				a.ledColor <- colors.Red
				go func() {
					time.Sleep(2 * time.Second)
					a.ledColor <- colors.Black
				}()
				continue
			}

			a.currentFile = fname

			if atomic.LoadInt32(&a.status) == paused {
				// send a PAUSE command to resume playback
				stdin.Write([]byte("PAUSE\n"))
				atomic.StoreInt32(&a.status, playing)
			}

			stdin.Write([]byte(fmt.Sprintf("LOAD %s\n", fname)))
			stdin.Write([]byte(fmt.Sprintf("GAIN %d\n", a.volume)))
			a.ledColor <- colors.Green
		case ctrlCmd := <-a.ctrl:
			switch ctrlCmd {
			case Pause:
				stdin.Write([]byte("PAUSE\n"))
				if a.currentFile == "" {
					atomic.StoreInt32(&a.status, stopped)
				} else {
					atomic.StoreInt32(&a.status, paused)
				}
				a.ledColor <- colors.Blue
			case Resume:
				stdin.Write([]byte("PAUSE\n"))
				if a.currentFile == "" {
					atomic.StoreInt32(&a.status, stopped)
					a.ledColor <- colors.Blue
				} else {
					atomic.StoreInt32(&a.status, playing)
					a.ledColor <- colors.Green
				}
			case Stop:
				a.stop(stdin)
			case Volume:
				// find the next volume
				for i, v := range volumes {
					if v == a.volume {
						a.volume = volumes[(i+1)%len(volumes)]
						break
					}
				}
				log.Printf("Volume set to %d", a.volume)
				stdin.Write([]byte(fmt.Sprintf("GAIN %d\n", a.volume)))
				// a.ledColor <- colors.Brown
			default:
				log.Println("Unknown control command:", ctrlCmd)
			}
		}
	}
}

func (a *Audio) stop(stdin io.WriteCloser) {
	stdin.Write([]byte("STOP\n"))
	a.currentFile = ""
	atomic.StoreInt32(&a.status, stopped)
	a.ledColor <- colors.Blue
}

func (a *Audio) cleanup() {
	if a.currentCmd == nil {
		return
	}
	a.currentCmd.Process.Kill()
	a.currentCmd = nil
}

func (a *Audio) setup() io.WriteCloser {

	a.currentCmd = exec.Command(a.cmd, "-R", "control")
	stdin, err := a.currentCmd.StdinPipe()
	if err != nil {
		log.Println("error opening stdin:", err)
		return nil
	}
	stdout, err := a.currentCmd.StdoutPipe()
	if err != nil {
		log.Println("error opening stdout:", err)
		return nil
	}

	if err := a.currentCmd.Start(); err != nil {
		log.Printf("error starting %s: %v", a.cmd, err)
		return nil
	}

	/* Read command output
	Specs from /usr/share/doc/mpg321/README.remote:

	@R MPG123
	mpg123 tagline. Output at startup.

	@I mp3-filename
	Prints out the filename of the mp3 file, minus the extension, or its ID3
	informations if available, in the form <title> <artist> <album>
	<year> <comment> <genre>.
	Happens after an mp3 file has been loaded.

	@S <a> <b> <c> <d> <e> <f> <g> <h> <i> <j> <k> <l>
	Outputs information about the mp3 file after loading.
	<a>: version of the mp3 file. Currently always 1.0 with madlib, but don't
		depend on it, particularly if you intend portability to mpg123 as well.
		Float/string.
	<b>: layer: 1, 2, or 3. Integer.
	<c>: Samplerate. Integer.
	<d>: Mode string. String.
	<e>: Mode extension. Integer.
	<f>: Bytes per frame (estimate, particularly if the stream is VBR). Integer.
	<g>: Number of channels (1 or 2, usually). Integer.
	<h>: Is stream copyrighted? (1 or 0). Integer.
	<i>: Is stream CRC protected? (1 or 0). Integer.
	<j>: Emphasis. Integer.
	<k>: Bitrate, in kbps. (i.e., 128.) Integer.
	<l>: Extension. Integer.

	@F <current-frame> <frames-remaining> <current-time> <time-remaining>
	Frame decoding status updates (once per frame).
	Current-frame and frames-remaining are integers; current-time and
	time-remaining floating point numbers with two decimal places.

	@P {0, 1, 2, 3}
	Stop/pause status.
	0 - Playing has stopped. When 'STOP' is entered, or the mp3 file is finished.
	1 - Playing is paused. Enter 'PAUSE' or 'P' to continue.
	2 - Playing has begun again.
	3 - Song has ended.

	@E <message>
	Unknown command or missing parameter.
	*/
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			out := scanner.Text()
			// read first 2 characters for command
			switch out[:2] {
			case "@P":
				switch strings.TrimSpace(out[2:]) {
				case "0":
					log.Println("Audio stopped")
					a.ledColor <- colors.Black
				case "1":
					log.Println("Audio paused")
					a.ledColor <- colors.Blue
				case "2":
					log.Println("Audio resumed")
					a.ledColor <- colors.Green
				case "3":
					log.Println("Audio finished")
					a.currentFile = ""
					a.ledColor <- colors.Black
				}
			}
		}
	}()

	return stdin
}
