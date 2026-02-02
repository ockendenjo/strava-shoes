package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

var colorCodeRed = "\033[91m"
var colorCodeReset = "\033[0m"

func main() {
	var isZipMode bool
	flag.BoolVar(&isZipMode, "zip", false, "zip")
	flag.Parse()

	cmd := exec.Command("find", "./cmd", "-type", "f", "-name", "main.go")
	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	if os.Getenv("NO_COLOR") != "" {
		colorCodeRed = ""
		colorCodeReset = ""
	}

	mainFiles := strings.Split(string(stdout), "\n")

	cmd = exec.Command("mkdir", "-p", "build")
	_, err = cmd.Output()
	if err != nil {
		panic(err)
	}

	hasError := false
	errorList := []string{}

	c := make(chan chanResult, len(mainFiles))
	parallelCount := 0
	remaining := 0
	maxParallel := getParallisation()
	fmt.Printf("Running build with parallisation: %d\n", maxParallel)

	build := func(file string) {
		go buildLambda(file, c, isZipMode)
		parallelCount++
		remaining++
	}
	okBuilds := make([]chanResult, 0, len(mainFiles))
	readChan := func() {
		chanRes := <-c
		remaining--
		if chanRes.err != nil {
			hasError = true
			errorList = append(errorList, chanRes.lambdaName)
		} else {
			okBuilds = append(okBuilds, chanRes)
		}
	}

	for i, file := range mainFiles {
		if len(file) < 1 {
			continue
		}

		build(file)

		if i == 0 || parallelCount > maxParallel {
			readChan()
		}
	}

	for remaining > 0 {
		readChan()
	}

	if hasError {
		l := log.New(os.Stderr, "", 0)
		l.Print(colorCodeRed)
		l.Printf("Lambda binary compilation failed for these main.go files:\n")
		for _, s := range errorList {
			l.Printf("   %s/main.go\n", s)
		}
		l.Println("See previous logging for error details")
		l.Print(colorCodeReset)
		os.Exit(1)
	}
	if isZipMode {
		printHashes(okBuilds)
	}
}

func printHashes(okBuilds []chanResult) {
	sort.Slice(okBuilds, func(i, j int) bool {
		return okBuilds[i].lambdaName < okBuilds[j].lambdaName
	})
	for _, build := range okBuilds {
		name := strings.Replace(build.lambdaName, "./cmd/", "", 1)
		fmt.Printf("%s %s\n", build.sha256, name)
	}
}

func getParallisation() int {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return 1
	}
	cpus := runtime.NumCPU()
	if cpus < 4 {
		//GitHub actions has only 2 CPUs and seems to be slower in parallel
		return 1
	}
	return cpus / 2
}

func buildLambda(mainFile string, c chan chanResult, isZipMode bool) {
	inputDir := getInputDirectory(mainFile)
	outPath := getOutputPath(mainFile)

	var sb strings.Builder

	cmd := exec.Command("go", "build", "-o", outPath, "-trimpath", "-buildvcs=false", "-ldflags=-w -s", inputDir) // #nosec G204 -- Subprocess needs to be launched with variable
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux")
	cmd.Env = append(cmd.Env, "GOARCH=arm64")
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	b := &bytes.Buffer{}
	_, _ = fmt.Fprintf(b, "%sBuild %s\n", colorCodeRed, inputDir)
	cmd.Stderr = b
	_, err := cmd.Output()
	if err != nil {
		l := log.New(os.Stderr, " ", 0)
		b.WriteString(colorCodeReset + "\n")
		l.Print(b.String())
		c <- chanResult{err: err, lambdaName: inputDir}
		return
	}

	sb.WriteString(fmt.Sprintf("Build %s\n", inputDir))
	size := getSize(outPath)
	if size > 0 {
		sb.WriteString(fmt.Sprintf("OK    %s %.1fMB\n", outPath, size))
	}

	var binHex string
	if isZipMode {
		err = buildZip(outPath)
		if err != nil {
			c <- chanResult{err: err, lambdaName: inputDir}
			return
		}

		err = os.Remove(outPath)
		if err != nil {
			c <- chanResult{err: err, lambdaName: inputDir}
			return
		}

		size = getSize(outPath + ".zip")
		if size > 0 {
			sb.WriteString(fmt.Sprintf("OK    %s %.1fMB\n", outPath+".zip", size))
		}

		binHex, err = getBinarySha256(outPath + ".zip")
		if err != nil {
			c <- chanResult{err: err, lambdaName: inputDir}
			return
		}
	}

	sb.WriteString("\n")
	fmt.Print(sb.String())
	c <- chanResult{err: nil, sha256: binHex, lambdaName: inputDir}
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

func buildZip(outputPath string) error {
	zipFile, err := os.Create(outputPath + ".zip")
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = addFileToZipDeterministic(zipWriter, outputPath)
	if err != nil {
		return err
	}

	return nil
}

func addFileToZipDeterministic(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Define a fixed timestamp (e.g., Unix epoch) to ensure determinism
	fixedTime := time.Date(time.Now().UTC().Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	// Create ZIP header
	header := &zip.FileHeader{
		Name:   "bootstrap",
		Method: zip.Deflate, // Use Deflate for compression
	}
	header.Modified = fixedTime // Set fixed modification time
	header.SetMode(0755)        // Ensure consistent file permissions

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

type chanResult struct {
	err        error
	lambdaName string
	sha256     string
}

func getSize(filePath string) float64 {
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0
	}

	size := float64(stat.Size()) / (1024 * 1024)
	return size
}

func getBinarySha256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
