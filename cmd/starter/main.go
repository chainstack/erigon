package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var ErrTooManyErrors = fmt.Errorf("Too many errors")

func log(format string, a ...any) {
	fmt.Printf(fmt.Sprintf("[starter] %s ", time.Now().Format("[01-02|15:04:05.000]"))+format+"\n", a...)
}

func process(cmd *exec.Cmd, errSubstring string, countLimit int, timeWindow time.Duration) error {

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log("Error creating stdout pipe: %v", err)
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log("Error creating stderr pipe: %v", err)
		return err
	}

	// async read from stdout and stderr
	// reader := io.MultiReader(stdout, stderr) // this will block
	r, w := io.Pipe()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := io.Copy(w, stdout)
		if err != nil {
			if err != io.EOF {
				log("Error copying stdout to pipe: %v", err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		_, err := io.Copy(w, stderr)
		if err != nil {
			if err != io.EOF {
				log("Error copying stderr to pipe: %v", err)
			}
		}
	}()
	go func() {
		wg.Wait()
		w.Close()
	}()

	reader := io.TeeReader(r, os.Stdout)

	if err := cmd.Start(); err != nil {
		log("Error starting process: %v", err)
		return err
	}

	lastOccurrences := []time.Time{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), errSubstring) && err != ErrTooManyErrors {
			log("Found substring: %s", errSubstring)
			lastOccurrences = append(lastOccurrences, time.Now())
			// Check if there are N or more occurrences
			if len(lastOccurrences) >= countLimit {
				first := lastOccurrences[0]
				last := lastOccurrences[len(lastOccurrences)-1]
				if last.Sub(first) < timeWindow {
					// Too many errors, terminate process
					// continue to read logs until process exits
					if err = cmd.Process.Signal(syscall.SIGINT); err != nil {
						log("Error sending signal to process: %v", err)
					}
					err = ErrTooManyErrors
				}
			}

		}
	}
	if err := scanner.Err(); err != nil {
		log("Error reading logs: %v", err)
		return err
	}
	return err
}

func main() {
	var (
		cmd *exec.Cmd
	)
	// Parse the command-line arguments
	flag.Parse()
	args := flag.Args()
	if len(args) < 4 {
		fmt.Println("Usage: go run main.go <substring> <countLimit> <timeWindow> <command> [<arg1> <arg2> ...]")
		return
	}

	errSubstr := args[0]
	countLimit, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("Error parsing error count limit: %v\nExample: 10, 100, 1000\n", err)
		return
	}

	timeWindow, err := time.ParseDuration(args[2])
	if err != nil {
		fmt.Printf("Error parsing error time window: %v\nExample: 10s, 1m, 1h\n", err)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log("Received signal %s for process %s %s %s %s %s", sig, os.Args[0], args[0], args[1], args[2], args[3])
		if cmd.Process != nil {
			if err := cmd.Process.Signal(sig); err != nil {
				log("Error sending signal to process: %v", err)
			}
		}
	}()

	for {
		log("Starting process %v", args[3:])
		cmd = exec.Command(args[3], args[4:]...) //nolint:gosec
		err := process(cmd, errSubstr, countLimit, timeWindow)
		if err != nil {
			switch err {
			case ErrTooManyErrors:
				log("Too many errors, restarting process")
				time.Sleep(time.Millisecond * 10)
				continue
			default:
				log("Process executing error: %v", err)
			}
		}
		log("Process stopped %v", args[3:])
		break
	}
	time.Sleep(time.Millisecond * 10)
}
