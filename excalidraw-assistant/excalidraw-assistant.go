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
	Type             string              `json:"type"`
	Version          int                 `json:"version"`
	Source           string              `json:"source"`
	Elements         []ExcalidrawElement `json:"elements"`
	AppState         map[string]interface{} `json:"appState"`
	Files            map[string]interface{} `json:"files"`
}

func main() {
	description := flag.String("description", "", "Natural language description of the diagram to create (required)")
	output := flag.String("output", "", "Output file path for .excalidraw JSON (required)")
	render := flag.Bool("render", false, "Also generate PNG output using the renderer")
	scale := flag.Int("scale", 2, "Render scale factor (default: 2)")
	width := flag.Int("width", 1920, "Render viewport width (default: 1920)")

	flag.Parse()

	if *description == "" {
		writeError(fmt.Errorf("--description is required"))
	}

	if *output == "" {
		writeError(fmt.Errorf("--output is required"))
	}

	// TODO: This is a placeholder implementation
	// In a full implementation, this would:
	// 1. Parse the description using an LLM or pattern matching
	// 2. Generate appropriate Excalidraw elements based on the description
	// 3. Apply colors from references/color-palette.md
	// 4. Use templates from references/element-templates.md
	
	// For now, create a simple example diagram
	diagram := createExampleDiagram(*description)

	// Write the .excalidraw JSON file
	jsonData, err := json.MarshalIndent(diagram, "", "  ")
	if err != nil {
		writeError(fmt.Errorf("failed to marshal JSON: %v", err))
	}

	if err := os.WriteFile(*output, jsonData, 0644); err != nil {
		writeError(fmt.Errorf("failed to write output file: %v", err))
	}

	result := fmt.Sprintf("Created Excalidraw diagram: %s", *output)

	// If render flag is set, call the Python renderer
	if *render {
		pngPath := strings.TrimSuffix(*output, filepath.Ext(*output)) + ".png"
		if err := renderDiagram(*output, pngPath, *scale, *width); err != nil {
			writeError(fmt.Errorf("failed to render diagram: %v", err))
		}
		result += fmt.Sprintf("\nRendered PNG: %s", pngPath)
	}

	writeOutput(result)
}

func createExampleDiagram(description string) ExcalidrawFile {
	// Create a simple example diagram
	// In a full implementation, this would generate elements based on the description
	
	return ExcalidrawFile{
		Type:    "excalidraw",
		Version: 2,
		Source:  "https://excalidraw.com",
		Elements: []ExcalidrawElement{
			{
				Type:            "rectangle",
				ID:              "elem1",
				X:               100,
				Y:               100,
				Width:           180,
				Height:          90,
				Angle:           0,
				StrokeColor:     "#1e3a5f",
				BackgroundColor: "#3b82f6",
				FillStyle:       "solid",
				StrokeWidth:     2,
				StrokeStyle:     "solid",
				Roughness:       0,
				Opacity:         100,
				GroupIds:        []string{},
				Seed:            12345,
				Version:         1,
				VersionNonce:    67890,
				IsDeleted:       false,
				BoundElements:   []map[string]string{{"id": "text1", "type": "text"}},
				Link:            nil,
				Locked:          false,
				Roundness:       map[string]interface{}{"type": 3},
			},
			{
				Type:            "text",
				ID:              "text1",
				X:               130,
				Y:               132,
				Width:           120,
				Height:          25,
				Angle:           0,
				StrokeColor:     "#1e3a5f",
				BackgroundColor: "transparent",
				FillStyle:       "solid",
				StrokeWidth:     1,
				StrokeStyle:     "solid",
				Roughness:       0,
				Opacity:         100,
				GroupIds:        []string{},
				Seed:            11111,
				Version:         1,
				VersionNonce:    22222,
				IsDeleted:       false,
				BoundElements:   nil,
				Link:            nil,
				Locked:          false,
				Text:            description,
				OriginalText:    description,
				FontSize:        16,
				FontFamily:      3,
				TextAlign:       "center",
				VerticalAlign:   "middle",
				ContainerID:     "elem1",
				LineHeight:      1.25,
			},
		},
		AppState: map[string]interface{}{
			"viewBackgroundColor": "#ffffff",
		},
		Files: map[string]interface{}{},
	}
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
