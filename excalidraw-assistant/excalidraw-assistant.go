package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExcalidrawElement struct {
	Type            string                 `json:"type"`
	ID              string                 `json:"id"`
	X               float64                `json:"x"`
	Y               float64                `json:"y"`
	Width           float64                `json:"width"`
	Height          float64                `json:"height"`
	Angle           float64                `json:"angle"`
	StrokeColor     string                 `json:"strokeColor"`
	BackgroundColor string                 `json:"backgroundColor"`
	FillStyle       string                 `json:"fillStyle"`
	StrokeWidth     int                    `json:"strokeWidth"`
	StrokeStyle     string                 `json:"strokeStyle"`
	Roughness       int                    `json:"roughness"`
	Opacity         int                    `json:"opacity"`
	GroupIds        []string               `json:"groupIds"`
	Seed            int                    `json:"seed"`
	Version         int                    `json:"version"`
	VersionNonce    int                    `json:"versionNonce"`
	IsDeleted       bool                   `json:"isDeleted"`
	BoundElements   interface{}            `json:"boundElements"`
	Link            interface{}            `json:"link"`
	Locked          bool                   `json:"locked"`
	Text            string                 `json:"text,omitempty"`
	OriginalText    string                 `json:"originalText,omitempty"`
	FontSize        int                    `json:"fontSize,omitempty"`
	FontFamily      int                    `json:"fontFamily,omitempty"`
	TextAlign       string                 `json:"textAlign,omitempty"`
	VerticalAlign   string                 `json:"verticalAlign,omitempty"`
	ContainerID     interface{}            `json:"containerId,omitempty"`
	LineHeight      float64                `json:"lineHeight,omitempty"`
	Points          [][]float64            `json:"points,omitempty"`
	StartBinding    interface{}            `json:"startBinding,omitempty"`
	EndBinding      interface{}            `json:"endBinding,omitempty"`
	StartArrowhead  interface{}            `json:"startArrowhead,omitempty"`
	EndArrowhead    interface{}            `json:"endArrowhead,omitempty"`
	Roundness       map[string]interface{} `json:"roundness,omitempty"`
}

type ExcalidrawFile struct {
	Type     string                 `json:"type"`
	Version  int                    `json:"version"`
	Source   string                 `json:"source"`
	Elements []ExcalidrawElement    `json:"elements"`
	AppState map[string]interface{} `json:"appState"`
	Files    map[string]interface{} `json:"files"`
}

func main() {
	input := flag.String("input", "", "Path to .excalidraw JSON file to validate/render (required)")
	output := flag.String("output", "", "Output PNG file path (defaults to input name with .png extension)")
	validateOnly := flag.Bool("validate-only", false, "Only validate JSON without rendering")
	scale := flag.Int("scale", 2, "Render scale factor (default: 2)")
	width := flag.Int("width", 1920, "Render viewport width (default: 1920)")

	flag.Parse()

	if *input == "" {
		writeError(fmt.Errorf("--input is required"))
	}

	// Default output to input name with .png extension
	pngOutput := *output
	if pngOutput == "" {
		pngOutput = strings.TrimSuffix(*input, ".excalidraw") + ".png"
	}

	// Read and validate the input JSON
	jsonData, err := os.ReadFile(*input)
	if err != nil {
		writeError(fmt.Errorf("failed to read input file: %v", err))
	}

	// Parse JSON to validate structure
	var diagram ExcalidrawFile
	if err := json.Unmarshal(jsonData, &diagram); err != nil {
		writeError(fmt.Errorf("invalid JSON: %v", err))
	}

	// Basic validation
	if diagram.Type != "excalidraw" {
		writeError(fmt.Errorf("invalid diagram type (expected 'excalidraw', got '%s')", diagram.Type))
	}

	if len(diagram.Elements) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: Diagram has no elements\n")
	}

	result := fmt.Sprintf("✓ JSON validated successfully: %d elements", len(diagram.Elements))

	// If validate-only flag is set, exit here
	if *validateOnly {
		writeOutput(result)
		return
	}

	// Render to PNG
	if err := renderDiagram(*input, pngOutput, *scale, *width); err != nil {
		writeError(fmt.Errorf("failed to render PNG: %v", err))
	}

	result += fmt.Sprintf("\n✓ Rendered PNG: %s", pngOutput)
	writeOutput(result)
}

func renderDiagram(excalidrawPath, pngPath string, scale, width int) error {
	// Get the tool directory to find the renderer
	toolDir, err := getToolDir()
	if err != nil {
		return err
	}

	rendererPath := filepath.Join(toolDir, "references", "render_excalidraw.py")

	// Check if renderer exists
	if _, err := os.Stat(rendererPath); os.IsNotExist(err) {
		return fmt.Errorf("renderer not found at %s. Run 'cd %s/references && uv sync && uv run playwright install chromium' to set up the renderer", rendererPath, toolDir)
	}

	// Call the Python renderer
	cmd := exec.Command("uv", "run", "python", rendererPath, excalidrawPath, "--output", pngPath, "--scale", fmt.Sprintf("%d", scale), "--width", fmt.Sprintf("%d", width))
	cmd.Dir = filepath.Join(toolDir, "references")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("renderer failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func getToolDir() (string, error) {
	// Get the directory where this binary is located
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	return filepath.Dir(execPath), nil
}

func writeOutput(result string) {
	fmt.Println(result)
}

func writeError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
