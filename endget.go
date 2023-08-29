package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

func installProgram(program string) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/RealLava/endget/main/applications/%s", program)
	fmt.Printf("Installing program '%s' from URL: %s\n", program, url)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching script:", err)
		return
	}
	defer response.Body.Close()

	scriptContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading script content:", err)
		return
	}

	cmd := exec.Command("bash", "-c", string(scriptContent))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
		return
	}
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

	installProgram(program)
}