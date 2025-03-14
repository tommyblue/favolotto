package favolotto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

type Audio struct {
	in        <-chan string
	ctrl      <-chan string
	storePath string

	currentFile string
	currentCmd  *exec.Cmd
}

// List of the possible control commands
var (
	Pause  = "pause"
	Resume = "resume"
	Stop   = "stop"
)

func NewAudio(storePath string, in <-chan string, ctrl <-chan string) *Audio {
	return &Audio{
		in:        in,
		ctrl:      ctrl,
		storePath: storePath,
	}
}

func (a *Audio) Run(ctx context.Context) {
	stdin := a.setup()
	for {
		select {
		case <-ctx.Done():
			a.cleanup()
			return
		case fname := <-a.in:
			// if the same file is requested, do nothing
			if a.currentFile == fname {
				break
			}
			fmt.Println("Audio file requested: ", fname)

			// check if the file exists
			if _, err := os.Stat(fname); os.IsNotExist(err) {
				log.Printf("File %s does not exist", fname)
				continue
			}

			a.currentFile = fname

			stdin.Write([]byte(fmt.Sprintf("LOAD %s\n", fname)))
			stdin.Write([]byte("GAIN 100\n"))
		case ctrlCmd := <-a.ctrl:
			switch ctrlCmd {
			case Pause:
				stdin.Write([]byte("PAUSE\n"))
			case Resume:
				stdin.Write([]byte("PAUSE\n"))
			case Stop:
				stdin.Write([]byte("STOP\n"))
				a.currentFile = ""
			default:
				fmt.Println("Unknown control command:", ctrlCmd)
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
	a.currentCmd = exec.Command("mpg321", "-R", "control")
	stdin, err := a.currentCmd.StdinPipe()
	if err != nil {
		fmt.Println("Errore nell'aprire stdin:", err)
		return nil
	}
	stdout, err := a.currentCmd.StdoutPipe()
	if err != nil {
		fmt.Println("Errore nell'aprire stdout:", err)
		return nil
	}

	if err := a.currentCmd.Start(); err != nil {
		fmt.Println("Errore nell'avviare mpg321:", err)
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
