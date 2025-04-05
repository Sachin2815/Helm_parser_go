The error you're seeing:

```
go: cannot find GOROOT directory: 'go' binary is trimmed and GOROOT is not set
```

means that your Go environment is either **not properly installed** or **not correctly configured**.

---

### ✅ Steps to Fix:

#### 1. **Check if Go is installed**
Run:
```bash
which go
```

If you see **no output**, Go is not installed.

---

#### 2. **Install or Reinstall Go**

On **RHEL/CentOS**, you can install Go like this:

```bash
sudo dnf install golang -y
```

Or, if you're using **Ubuntu/Debian**:

```bash
sudo apt install golang -y
```

Alternatively, **manual method** (recommended for the latest version):

```bash
# 1. Download latest Go (adjust version as needed)
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz

# 2. Remove any existing Go
sudo rm -rf /usr/local/go

# 3. Extract to /usr/local
sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
```

---

#### 3. **Set Go Environment Variables**

Edit your shell profile:

```bash
nano ~/.bashrc
```

Add at the end:

```bash
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

Then **apply the changes**:

```bash
source ~/.bashrc
```

---

#### 4. **Verify Installation**

```bash
go version
go env
```

You should now see Go version and properly set `GOROOT` and `GOPATH`.

---

### ✅ Then try again:

```bash
go run main.go
```

Let me know if you'd like help with installing via the manual method or setting it up for your RHEL VM.

This error means your Go project is **missing a `go.mod` file**, which is required for dependency management (like `gopkg.in/yaml.v2`).

---

### ✅ Fix Steps

#### 1. **Initialize the Go module**
Run this from the root of your project (`helm_parser_go`):

```bash
go mod init Helm_parser_go  ##Run this command after the installaton
```

This will create a `go.mod` file.

---

#### 2. **Download the dependencies**
Now run:

```bash
go mod tidy
```

This will download all the required modules, including:

```bash
gopkg.in/yaml.v2
```

---

#### 3. **Then run your project again**
```bash
go run main.go
```

---

### 🧠 Summary
- `go mod init <module_name>` creates a `go.mod` file for your project.
- `go mod tidy` resolves and downloads all dependencies.
- Without `go.mod`, Go doesn’t know how to pull external packages.

Let me know if you want me to generate a sample `go.mod` file based on your `main.go`.

Ah, Go is strict about unused imports — it won’t compile if there's even one unused package.

You're getting:
```
"io/fs" imported and not used
```

So, to fix it:

---

### ✅ **Remove the unused import**
In your `main.go`, find and **delete** or **comment** this line:
```go
"io/fs"
```

You likely have this block:
```go
import (
	"fmt"
	"io/fs" // ❌ remove this line
	"os"
	"path/filepath"
)
```

---

### After fixing:
Run again:
```bash
go run main.go
```

And you should see:
```
🚀 Server started at http://localhost:8080
```

Let me know once it's running — we can test the form after that.

