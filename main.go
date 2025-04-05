package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// 🧱 Structs
type ImageInfo struct {
	Name   string
	Size   string
	Layers int
}

type ChartData struct {
	AppVersion string `yaml:"appVersion"`
}

type ValuesData struct {
	Image struct {
		Repository string `yaml:"repository"`
		Tag        string `yaml:"tag"`
	} `yaml:"image"`
}

// 📅 Generate timestamp in YYYYMMDDHHMMSS format
func generateTimestamp() string {
	return time.Now().Format("20060102150405")
}

// 🔄 Clone a GitHub Helm repo into repo_db/<timestamp>_<repo-name>
func cloneHelmRepo(repoURL string) (string, error) {
	parts := strings.Split(strings.TrimSuffix(repoURL, "/"), "/")
	repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")
	timestamp := generateTimestamp()
	targetDir := filepath.Join("repo_db", fmt.Sprintf("%s_%s", timestamp, repoName))

	if err := os.MkdirAll("repo_db", os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create repo_db directory: %v", err)
	}

	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
	}

	fmt.Printf("✅ Repository cloned to: %s\n", targetDir)
	return targetDir, nil
}

// 📂 Navigate to /charts and list Helm subdirectories
func navigateToHelmChart(repoPath string) ([]string, error) {
	chartsDir := filepath.Join(repoPath, "charts")
	info, err := os.Stat(chartsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Println("⚠️ No 'charts' directory found!")
		return []string{}, nil
	}

	var helmChartDirs []string
	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(chartsDir, entry.Name())
			fmt.Println("🔍 Found Helm chart directory:", dirPath)
			helmChartDirs = append(helmChartDirs, dirPath)
		}
	}

	return helmChartDirs, nil
}

// 🔍 Parse values.yaml and Chart.yaml to construct image name
func checkAndParseHelmFiles(helmChartDirs []string) (string, error) {
	for _, chartDir := range helmChartDirs {
		fmt.Println("🔎 Checking:", chartDir)

		chartYamlPath := filepath.Join(chartDir, "Chart.yaml")
		valuesYamlPath := filepath.Join(chartDir, "values.yaml")

		if _, err := os.Stat(chartYamlPath); os.IsNotExist(err) {
			fmt.Println("⚠️ Chart.yaml not found in:", chartDir)
			continue
		}
		if _, err := os.Stat(valuesYamlPath); os.IsNotExist(err) {
			fmt.Println("⚠️ values.yaml not found in:", chartDir)
			continue
		}

		chartContent, err := ioutil.ReadFile(chartYamlPath)
		if err != nil {
			return "", err
		}
		valuesContent, err := ioutil.ReadFile(valuesYamlPath)
		if err != nil {
			return "", err
		}

		var chartData ChartData
		var valuesData ValuesData

		yaml.Unmarshal(chartContent, &chartData)
		yaml.Unmarshal(valuesContent, &valuesData)

		repo := valuesData.Image.Repository
		tag := valuesData.Image.Tag
		if repo == "" {
			fmt.Println("⚠️ No image repository found in values.yaml")
			continue
		}

		if tag == "" {
			tag = chartData.AppVersion
		}

		imageName := fmt.Sprintf("%s:%s", repo, tag)
		fmt.Println("📦 Image to inspect:", imageName)
		return imageName, nil
	}

	return "", fmt.Errorf("no valid Helm chart found")
}

// 🛠️ Utility to run shell commands
func runCommand(cmd []string) (string, error) {
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	return string(out), err
}

func getImageSizeAndLayers(imageName string) (*ImageInfo, error) {
	// 🔍 Check if image exists locally
	imageID, _ := runCommand([]string{"docker", "images", "-q", imageName})
	if imageID == "" {
		fmt.Println("⬇️ Pulling image:", imageName)
		cmd := exec.Command("docker", "pull", imageName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to pull image: %v", err)
		}
	} else {
		fmt.Printf("✅ Image '%s' already exists locally.\n", imageName)
	}

	// 🔍 Inspect image
	inspectOut, err := runCommand([]string{"docker", "image", "inspect", imageName})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %v", err)
	}

	var inspectData []map[string]interface{}
	err = json.Unmarshal([]byte(inspectOut), &inspectData)
	if err != nil || len(inspectData) == 0 {
		return nil, fmt.Errorf("failed to parse inspect data")
	}

	// 📦 Parse size and layers
	size := int(inspectData[0]["Size"].(float64))
	layers := len(inspectData[0]["RootFS"].(map[string]interface{})["Layers"].([]interface{}))

	// 📤 Print info like Python version
	fmt.Printf("📦 Image: %s\n", imageName)
	fmt.Printf("   🔹 Size: %.2f MB\n", float64(size)/(1000*1000))
	fmt.Printf("   🔹 Layers: %d\n", layers)

	return &ImageInfo{
		Name:   imageName,
		Size:   fmt.Sprintf("%.2f MB", float64(size)/(1000*1000)),
		Layers: layers,
	}, nil
}


// 🌐 Render Home Page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

// 🔍 Handle /imagedetails request
func imageDetailsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	repoURL := r.FormValue("repo_url")

	targetDir, err := cloneHelmRepo(repoURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helmDirs, err := navigateToHelmChart(targetDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	imageName, err := checkAndParseHelmFiles(helmDirs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	imageInfo, err := getImageSizeAndLayers(imageName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/image_details.html"))
	tmpl.Execute(w, imageInfo)
}

// 🚀 Start server
func main() {
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/imagedetails", imageDetailsHandler)

	fmt.Println("🚀 Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
