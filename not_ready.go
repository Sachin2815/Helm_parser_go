package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)



// 🧱 Define the struct here (just below imports)
type ImageInfo struct {
	Name   string
	Size   string
	Layers int
}

// generateTimestamp returns a timestamp in YYYYMMDDHHMMSS format
func generateTimestamp() string {
	return time.Now().Format("20060102150405")
}

// cloneHelmRepo clones a GitHub repo into a timestamped directory inside repo_db
func cloneHelmRepo(repoURL string) (string, error) {
	// Extract the repo name
	parts := strings.Split(strings.TrimSuffix(repoURL, "/"), "/")
	repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")

	// Generate target directory
	timestamp := generateTimestamp()
	targetDir := filepath.Join("repo_db", fmt.Sprintf("%s_%s", timestamp, repoName))

	// Ensure parent directory exists
	if err := os.MkdirAll("repo_db", os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create repo_db directory: %v", err)
	}

	// Clone the repo
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
	}

	fmt.Printf("Repository cloned to: %s\n", targetDir)
	return targetDir, nil
}


// package main



// navigateToHelmChart returns a slice of directories inside the charts/ folder of the repo
func navigateToHelmChart(repoPath string) ([]string, error) {
	chartsDir := filepath.Join(repoPath, "charts")

	// Check if charts directory exists
	info, err := os.Stat(chartsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Println("No 'charts' directory found!")
		return []string{}, nil
	}

	var helmChartDirs []string

	// Read contents of the charts directory
	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(chartsDir, entry.Name())
			fmt.Println("Found Helm chart directory:", dirPath)
			helmChartDirs = append(helmChartDirs, dirPath)
		}
	}

	fmt.Println("Helm chart directories found:", helmChartDirs)
	return helmChartDirs, nil
}

// package main



type ChartData struct {
	AppVersion string `yaml:"appVersion"`
}

type ValuesData struct {
	Image struct {
		Repository string `yaml:"repository"`
		Tag        string `yaml:"tag"`
	} `yaml:"image"`
}

type ImageInfo struct {
	Name   string
	Size   string
	Layers int
}

// Check and parse Helm files in given directories
func checkAndParseHelmFiles(helmChartDirs []string) (string, error) {
	for _, chartDir := range helmChartDirs {
		fmt.Println("Currently in:", chartDir)
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

		fmt.Println("✅ Found Chart.yaml and values.yaml in:", chartDir)

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
			fmt.Println("No image repository found in values.yaml")
			continue
		}

		if tag == "" {
			tag = chartData.AppVersion
		}
		imageName := fmt.Sprintf("%s:%s", repo, tag)
		fmt.Println("Your image is:", imageName)
		return imageName, nil
	}
	return "", fmt.Errorf("no valid Helm chart found")
}

func runCommand(cmd []string) (string, error) {
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	return string(out), err
}

func getImageSizeAndLayers(imageName string) (*ImageInfo, error) {
	imageID, _ := runCommand([]string{"docker", "images", "-q", imageName})
	if imageID == "" {
		fmt.Println("⬇️ Pulling image:", imageName)
		_, err := runCommand([]string{"docker", "pull", imageName})
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("✅ Image already exists locally:", imageName)
	}

	inspectOut, err := runCommand([]string{"docker", "image", "inspect", imageName})
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	err = json.Unmarshal([]byte(inspectOut), &inspectData)
	if err != nil || len(inspectData) == 0 {
		return nil, fmt.Errorf("failed to parse inspect data")
	}

	size := int(inspectData[0]["Size"].(float64))
	layers := len(inspectData[0]["RootFS"].(map[string]interface{})["Layers"].([]interface{}))

	imageInfo := &ImageInfo{
		Name:   imageName,
		Size:   fmt.Sprintf("%.2f MB", float64(size)/(1000*1000)),
		Layers: layers,
	}

	fmt.Println("📦 Image:", imageInfo.Name)
	fmt.Println("   🔹 Size:", imageInfo.Size)
	fmt.Println("   🔹 Layers:", imageInfo.Layers)

	return imageInfo, nil
}




// func main() {
// 	repoURL := "https://github.com/user/repo.git"
// 	targetDir, err := cloneHelmRepo(repoURL)
// 	if err != nil {
// 		fmt.Println("❌ Error:", err)
// 		return
// 	}

// 	fmt.Println("✅ Cloned repo path:", targetDir)
// }

// func main() {
// 	repoPath := "repo_db/20250403153000_myrepo"
// 	dirs, err := navigateToHelmChart(repoPath)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	if len(dirs) == 0 {
// 		fmt.Println("No Helm charts found.")
// 	} else {
// 		fmt.Println("✅ Helm charts found:", dirs)
// 	}
// }
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

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

func main() {
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/imagedetails", imageDetailsHandler)

	fmt.Println("🚀 Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
