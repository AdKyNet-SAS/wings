package router

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pterodactyl/wings/router/middleware"
	"github.com/pterodactyl/wings/server/filesystem"
)

// fileSHA256 returns the hex encoded SHA256 hash of the file at the given path,
// resolved within the server's filesystem root.
func fileSHA256(fs *filesystem.Filesystem, p string) (string, error) {
	f, _, err := fs.File(p)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// get installed version
func getInstalledVersion(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	fs := s.Filesystem()

	// get first environment variable that includes SERVER and .jar in the name/value
	variables := s.GetEnvironmentVariables()
	var jar string
	for _, variable := range variables {
		if strings.Contains(variable, "JAR") && strings.Contains(variable, ".jar") {
			parts := strings.SplitN(variable, "=", 2)
			if len(parts) == 2 {
				jar = parts[1]
			}

			break
		}
	}

	if jar == "" {
		jar = "server.jar"
	}

	// check if libraries/net/minecraftforge/forge folder exists
	if _, err := fs.Stat("libraries/net/minecraftforge/forge"); err == nil {
		// get first folder in libraries/net/minecraftforge/forge
		files, err := fs.ListDirectory("libraries/net/minecraftforge/forge")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load files."})
			return
		}

		if len(files) > 0 {
			folder := files[0].Name()

			// get hash of -server.jar or -universal.jar inside the first folder
			files, err := fs.ListDirectory(path.Join("libraries/net/minecraftforge/forge", folder))

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load files."})
				return
			}

			for _, file := range files {
				if strings.HasSuffix(file.Name(), "-server.jar") || strings.HasSuffix(file.Name(), "-universal.jar") {
					hash, err := fileSHA256(fs, path.Join("libraries/net/minecraftforge/forge", folder, file.Name()))

					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get hash."})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "hash": hash})
					return
				}
			}
		}
	}

	// check if libraries/net/neoforged/neoforge folder exists
	if _, err := fs.Stat("libraries/net/neoforged/neoforge"); err == nil {
		// get first folder in libraries/net/neoforged/neoforge
		files, err := fs.ListDirectory("libraries/net/neoforged/neoforge")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load files."})
			return
		}

		if len(files) > 0 {
			folder := files[0].Name()

			// get hash of -server.jar or -universal.jar inside the first folder
			files, err := fs.ListDirectory(path.Join("libraries/net/neoforged/neoforge", folder))

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load files."})
				return
			}

			for _, file := range files {
				if strings.HasSuffix(file.Name(), "-server.jar") || strings.HasSuffix(file.Name(), "-universal.jar") {
					hash, err := fileSHA256(fs, path.Join("libraries/net/neoforged/neoforge", folder, file.Name()))

					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get hash."})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "hash": hash})
					return
				}
			}
		}
	}

	// hash server.jar, if it exists
	if _, err := fs.Stat(jar); err == nil {
		hash, err := fileSHA256(fs, jar)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get hash."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "hash": hash})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": false, "error": "No version found."})
}
