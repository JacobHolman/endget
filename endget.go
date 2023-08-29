package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const repositoryURL = "https://api.github.com/repos/RealLava/endget/contents/applications"

type GitHubFile struct {
	Name string `json:"name"`
}

func fetchAvailablePrograms() ([]string, error) {
	resp, err := http.Get(repositoryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %s", resp.Status)
	}

	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var programs []string
	for _, file := range files {
		program := strings.TrimSuffix(file.Name, ".sh")
		programs = append(programs, program)
	}

	return programs, nil
}

const progressBarWidth = 50

func displayProgressBar(progress float64) {
	bar := "["
	completeWidth := int(progress * float64(progressBarWidth))
	for i := 0; i < progressBarWidth; i++ {
		if i < completeWidth {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"
	fmt.Printf("\r%s %.1f%%", bar, progress*100)
}

func installProgram(program string, done chan bool) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/RealLava/endget/main/applications/%s", program)
	fmt.Printf("Installing program '%s' from URL: %s\n", program, url)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching script:", err)
		done <- false
		return
	}
	defer response.Body.Close()

	scriptContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading script content:", err)
		done <- false
		return
	}

	cmd := exec.Command("bash", "-c", string(scriptContent))
	outputPipe, _ := cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err)
		done <- false
		return
	}

	totalBytes := int(response.ContentLength)
	buf := make([]byte, 1024)
	bytesRead := 0

	go func() {
		for {
			n, err := outputPipe.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading output:", err)
				break
			}
			bytesRead += n
			displayProgressBar(float64(bytesRead) / float64(totalBytes))
			os.Stdout.Write(buf[:n])
		}
	}()

	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error:", err)
		done <- false
		return
	}

	fmt.Println()
	done <- true
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: endget install <program>")
		return
	}

	command := os.Args[1]
	if command != "install" {
		fmt.Println("Usage: endget install <program>")
		return
	}

	programs, err := fetchAvailablePrograms()
	if err != nil {
		fmt.Println("Error fetching available programs:", err)
		return
	}

	program := os.Args[2]

	found := false
	for _, p := range programs {
		if p == program {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Program '%s' not found in the repository.\n", program)
		return
	}

	done := make(chan bool)
	go installProgram(program, done)

	go func() {
		for i := 0; i <= progressBarWidth; i++ {
			time.Sleep(100 * time.Millisecond)
			displayProgressBar(float64(i) / float64(progressBarWidth))
		}
		fmt.Println()
		done <- true
	}()

	installationResult := <-done
	if installationResult {
		fmt.Println("Installation completed successfully.")
	} else {
		fmt.Println("Installation failed.")
	}
}
