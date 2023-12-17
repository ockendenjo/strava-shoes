package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	cmd := exec.Command("find", "./cmd", "-type", "f", "-name", "main.go")
	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	mainFiles := strings.Split(string(stdout), "\n")

	cmd = exec.Command("mkdir", "-p", "build")
	_, err = cmd.Output()
	if err != nil {
		panic(err)
	}

	hasError := false

	c := make(chan chanResult, len(mainFiles))
	parallelCount := 0
	remaining := 0
	maxParallel := getParallisation()
	fmt.Printf("Running build with parallisation: %d\n", maxParallel)

	for _, file := range mainFiles {
		if len(file) < 1 {
			continue
		}

		go buildLambda(file, c)
		parallelCount++
		remaining++

		if parallelCount > maxParallel {
			chanRes := <-c
			remaining--
			if chanRes.err != nil {
				hasError = true
			}
		}
	}

	for k := 0; k < remaining; k++ {
		chanRes := <-c
		if chanRes.err != nil {
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}

func getParallisation() int {
	cpus := runtime.NumCPU()
	if cpus < 4 {
		//GitHub actions has only 2 CPUs and seems to be slower in parallel
		return 1
	}
	return cpus
}

func buildLambda(mainFile string, c chan chanResult) {
	inputDir := getInputDirectory(mainFile)
	outPath := getOutputPath(mainFile)

	cmd := exec.Command("go", "build", "-o", outPath, "-trimpath", "-buildvcs=false", "-ldflags=-w -s", inputDir)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux")
	cmd.Env = append(cmd.Env, "GOARCH=arm64")
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()
	if err != nil {
		c <- chanResult{err: err}
		return
	}

	size := float64(0)
	fi, err := os.Stat(outPath)
	if err == nil {
		size = float64(fi.Size()) / (1000 * 1000)
		fmt.Printf("Build %s\nOK    %s %.1fMB\n\n", inputDir, outPath, size)
	} else {
		fmt.Printf("Build %s\nOK    %s\n\n", inputDir, outPath)
	}
	c <- chanResult{err: nil}
}

func getInputDirectory(mainFile string) string {
	return strings.Replace(mainFile, "/main.go", "", 1)
}

// getOutputPath flattens the directory structure replacing `/` with `-` and sets the correct output directory
func getOutputPath(mainFile string) string {
	outDir := strings.Replace(mainFile, "/main.go", "", 1)
	outDir = strings.Replace(outDir, "./cmd/", "", 1)
	outDir = strings.ReplaceAll(outDir, "/", "-")

	return fmt.Sprintf("build/%s/bootstrap", outDir)
}

type chanResult struct {
	err error
}
