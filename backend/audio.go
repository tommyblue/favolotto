package favolotto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

type Audio struct {
	in        <-chan string
	ctrl      <-chan string
	ledColor  chan<- string
	storePath string

	cmd         string
	currentFile string
	currentCmd  *exec.Cmd
}

// List of the possible control commands
var (
	Pause  = "pause"
	Resume = "resume"
	Stop   = "stop"
)

func NewAudio(storePath string, in <-chan string, ctrl <-chan string, ledColor chan<- string) (*Audio, error) {
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
			// if the same file is requested, do nothing
			// TODO: if the file is the same but the audio is paused, resume it
			if a.currentFile == fname {
				log.Println("Same audio file requested")
				a.ledColor <- "orange"
				go func() {
					time.Sleep(2 * time.Second)
					a.ledColor <- "black"
				}()
				break
			}
			log.Println("Audio file requested: ", fname)

			// check if the file exists
			if _, err := os.Stat(fname); os.IsNotExist(err) {
				log.Printf("File %s does not exist", fname)
				a.ledColor <- "red"
				go func() {
					time.Sleep(2 * time.Second)
					a.ledColor <- "black"
				}()
				continue
			}

			a.currentFile = fname

			stdin.Write([]byte(fmt.Sprintf("LOAD %s\n", fname)))
			stdin.Write([]byte("GAIN 100\n"))
			a.ledColor <- "green"
		case ctrlCmd := <-a.ctrl:
			switch ctrlCmd {
			case Pause:
				stdin.Write([]byte("PAUSE\n"))
				a.ledColor <- "blue"
			case Resume:
				stdin.Write([]byte("PAUSE\n"))
				a.ledColor <- "green"
			case Stop:
				stdin.Write([]byte("STOP\n"))
				a.currentFile = ""
				a.ledColor <- "blue"
			default:
				log.Println("Unknown control command:", ctrlCmd)
			}
		}
	}
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
		log.Println("error starting mpg321:", err)
		return nil
	}

	// Lettura output per debug
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text()) // Debug output mpg321
		}
	}()

	return stdin
}
