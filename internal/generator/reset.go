package generator

import (
	"bufio"
	"os"
	"path"
	"strings"
)

// deleteGenFiles deletes files generated previously by the same generator
func (t *Transport) deleteGenFiles(dirPath, genFilePrefix string) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		t.log.Warn().Err(err).Msg("reset gen files")
		return
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !strings.HasSuffix(dirEntry.Name(), ".go") {
			continue
		}
		// file with .go extension
		filePath := path.Join(dirPath, dirEntry.Name())
		t.deleteGenFile(filePath, genFilePrefix)
	}
}

func (t *Transport) deleteGenFile(filePath string, genFilePrefix string) {
	file, err := os.Open(filePath)
	if err != nil {
		t.log.Error().Err(err).Str("filePath", filePath).Msg("failed to open .go file")
		return
	}
	defer file.Close()

	firstLine, err := bufio.NewReader(file).ReadString('\n')
	if err != nil {
		t.log.Error().Err(err).Msg("failed to read from .go file")
		return
	}
	firstLineComment := strings.TrimSpace(strings.TrimPrefix(firstLine, "//"))
	if !strings.HasPrefix(firstLineComment, genFilePrefix) {
		return
	}
	// this is indeed a generated file, delete it
	if err = os.Remove(filePath); err != nil {
		t.log.Warn().Err(err).Msg("failed to clean up the old gen files")
	}
}
