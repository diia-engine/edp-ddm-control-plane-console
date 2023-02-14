package registry

import (
	"fmt"
	"os"

	"github.com/patrickmn/go-cache"
)

type CachedFile struct {
	RepoPath string
	TempPath string
}

func CachedFilesIndex(registryName string) string {
	return fmt.Sprintf("%s-cached-files", registryName)
}

func CacheRepoFiles(tempFolder, registryName string, repoFiles map[string]string, c *cache.Cache) error {
	cachedFiles := make([]CachedFile, 0, len(repoFiles))

	for p, content := range repoFiles {
		fp, err := os.CreateTemp(tempFolder, registryName)
		if err != nil {
			return fmt.Errorf("unable to create file, %w", err)
		}

		if _, err := fp.WriteString(content); err != nil {
			return fmt.Errorf("unable to write string, %w", err)
		}

		if err := fp.Close(); err != nil {
			return fmt.Errorf("unable to close file, %w", err)
		}

		cachedFiles = append(cachedFiles, CachedFile{RepoPath: p, TempPath: fp.Name()})
	}

	if len(cachedFiles) > 0 {
		c.SetDefault(CachedFilesIndex(registryName), cachedFiles)
	}

	return nil
}
