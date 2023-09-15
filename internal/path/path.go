package path

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	goModPathByDirPath = make(map[string]string)
	pkgPathByGoModPath = make(map[string]string)
	modulePrefix       = []byte("\nmodule ")
)

func PkgPath(dirPath string) (string, error) {
	goModPath, err := goModPath(dirPath)
	if err != nil {
		return "", err
	}

	if strings.Contains(goModPath, "go.mod") {
		pkgPath, err := goModPathToPkgPath(dirPath, goModPath)
		if err != nil {
			return "", err
		}
		return pkgPath, nil
	}
	return "", errors.New("no go.mod found")
}

func goModPath(dirPath string) (string, error) {
	goMod, ok := goModPathByDirPath[dirPath]
	if ok {
		return goMod, nil
	}
	defer func() {
		goModPathByDirPath[dirPath] = goMod
	}()

	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Dir = dirPath
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// maybe iterate over parent directories until go mod path is retrieved

	return string(bytes.TrimSpace(stdout)), nil
}

func goModPathToPkgPath(dirPath, goModPath string) (string, error) {
	modulePath := modulePath(goModPath)

	if modulePath == "" {
		return "", fmt.Errorf("failed to find module path by the path to go mod file: %s", goModPath)
	}

	// example
	//
	// goModPath = "/home/user/myproject/go.mod"
	// dirPath = "/home/user/myproject/subdir"
	// modulePath = "github.com/myuser/myproject"
	//
	// 1. filepath.Dir(goModPath)          =>   "/home/user/myproject"
	// 2. strings.TrimPrefix(dirPath, ...) =>   "/subdir"
	// 3. filepath.ToSlash(...)            =>   OS-specific path separator to the forward slash
	// 4. path.Join(modulePath, ...)       =>   "github.com/myuser/myproject/subdir"
	moduleRootPath := filepath.Dir(goModPath)
	subdirPath := strings.TrimPrefix(dirPath, moduleRootPath)
	pkgPath := path.Join(modulePath, filepath.ToSlash(subdirPath))
	return pkgPath, nil
}

func modulePath(goModPath string) string {

	pkgPath, ok := pkgPathByGoModPath[goModPath]

	if ok {
		return pkgPath
	}

	defer func() {
		pkgPathByGoModPath[goModPath] = pkgPath
	}()

	data, err := os.ReadFile(goModPath)

	if err != nil {
		return ""
	}

	var i int

	if bytes.HasPrefix(data, modulePrefix[1:]) {
		i = 0
	} else {
		i = bytes.Index(data, modulePrefix)
		if i < 0 {
			return ""
		}
		i++
	}

	line := data[i:]

	// Cut line at \n, drop trailing \r if present.
	if j := bytes.IndexByte(line, '\n'); j >= 0 {
		line = line[:j]
	}

	if line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	line = line[len("module "):]

	// If quoted, unquote.
	pkgPath = strings.TrimSpace(string(line))

	if pkgPath != "" && pkgPath[0] == '"' {
		s, err := strconv.Unquote(pkgPath)
		if err != nil {
			return ""
		}
		pkgPath = s
	}
	return pkgPath
}
