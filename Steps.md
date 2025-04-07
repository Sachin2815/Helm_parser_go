# Helm Chart Image Inspector - Go Project

This project is a Go-based web application that accepts a GitHub Helm chart link, extracts Docker image information from it, pulls the image, and returns its size and number of layers.

---

## ✅ Steps I Followed

### 1. Setup Web Server and Form
- Created a Go HTTP server with two routes:
  - `/home` for input form
  - `/imagedetails` for processing the form and showing image data
- Used Go’s HTML templates to render basic UI.

### 2. Clone Helm Chart Repository
- Took GitHub repo URL as input from the user.
- Used `git clone` to copy the repo into a local timestamped folder.
- Ensured separation between different user requests.

### 3. Identify Helm Chart Directories
- Navigated to the `/charts/` directory inside the cloned repo.
- Listed all subfolders that could contain valid Helm charts.

### 4. Parse YAML Files
- Read `Chart.yaml` to extract `appVersion`.
- Read `values.yaml` to extract image `repository` and `tag`.
- If `tag` not found, used `appVersion` as fallback.

### 5. Construct Full Image Name
- Combined image repo and tag like `nginx:1.19`.
- Prepared it for inspection and pulling.

### 6. Pull and Inspect Docker Image
- Checked if the image exists locally.
- If not, used `docker pull` to download the image.
- Used `docker inspect` to get:
  - Total size of the image
  - Number of layers

### 7. Return Result to User
- Passed the image name, size, and layer count to a result HTML template.
- Displayed it clearly to the user via the browser.

---

## 🛠Tools Used

- **Go (Golang)** – backend logic and server
- **HTML Templates** – frontend rendering
- **Git CLI** – clone Helm repos
- **Docker CLI** – pull and inspect images
- **Helm Chart Structure** – parse and analyze

---

## 📦 Example Output

```json
{
  "Name": "nginx:1.19",
  "Size": "24.58 MB",
  "Layers": 7
}
