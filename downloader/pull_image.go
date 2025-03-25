package downloader

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PullImage pulls a Docker image from Docker Hub using the v1 API (similar to the bash script)
func PullImage(name, tag, destDir string) error {
	fmt.Printf("Pulling image %s:%s...\n", name, tag)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Get auth token
	fmt.Println("Getting authentication token...")
	token, err := getV1AuthToken(name)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %v", err)
	}

	registry := "https://registry-1.docker.io/v1"

	// Get image ID
	fmt.Println("Getting image ID...")
	imageID, err := getImageID(registry, token, name, tag)
	if err != nil {
		return fmt.Errorf("failed to get image ID: %v", err)
	}

	if len(imageID) != 64 {
		return fmt.Errorf("no image named '%s:%s' exists", name, tag)
	}

	// Get image ancestry
	fmt.Println("Getting image ancestry...")
	ancestry, err := getImageAncestry(registry, token, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image ancestry: %v", err)
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "docker-pull-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create rootfs directory
	rootfsDir := filepath.Join(destDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return fmt.Errorf("failed to create rootfs directory: %v", err)
	}

	// Download and extract layers
	fmt.Println("Downloading and extracting layers...")
	for i, id := range ancestry {
		fmt.Printf("Downloading layer %d of %d: %s\n", i+1, len(ancestry), id)

		layerURL := fmt.Sprintf("%s/images/%s/layer", registry, id)
		layerPath := filepath.Join(tempDir, "layer.tar")

		if err := downloadLayerWithAuth(layerURL, token, layerPath); err != nil {
			return fmt.Errorf("failed to download layer %s: %v", id, err)
		}

		fmt.Printf("Extracting layer %d to rootfs...\n", i+1)
		if err := extractTarLayer(layerPath, rootfsDir); err != nil {
			return fmt.Errorf("failed to extract layer %s: %v", id, err)
		}
	}

	// Save image source info
	imageInfo := fmt.Sprintf("%s:%s", name, tag)
	infoPath := filepath.Join(destDir, "image.json")

	info := map[string]string{
		"name": name,
		"tag":  tag,
	}
	infoData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal image info: %v", err)
	}

	if err := os.WriteFile(infoPath, infoData, 0644); err != nil {
		return fmt.Errorf("failed to write image info: %v", err)
	}

	fmt.Printf("Successfully pulled and extracted image %s:%s to %s\n", name, tag, destDir)
	return nil
}

// getV1AuthToken gets an authentication token for Docker Hub using the v1 API
func getV1AuthToken(repo string) (string, error) {
	url := fmt.Sprintf("https://index.docker.io/v1/repositories/%s/images", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Docker-Token", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request failed with status: %s", resp.Status)
	}

	return resp.Header.Get("X-Docker-Token"), nil
}

// getImageID gets the image ID for a specific tag
func getImageID(registry, token, repo, tag string) (string, error) {
	url := fmt.Sprintf("%s/repositories/%s/tags/%s", registry, repo, tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image ID request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Remove quotes from the response
	imageID := strings.Trim(string(body), "\"")
	return imageID, nil
}

// getImageAncestry gets the ancestry (layers) for an image
func getImageAncestry(registry, token, imageID string) ([]string, error) {
	url := fmt.Sprintf("%s/images/%s/ancestry", registry, imageID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ancestry request failed with status: %s", resp.Status)
	}

	var ancestry []string
	if err := json.NewDecoder(resp.Body).Decode(&ancestry); err != nil {
		return nil, err
	}

	return ancestry, nil
}

// downloadLayerWithAuth downloads a layer with authentication
func downloadLayerWithAuth(url, token, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("layer download failed with status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarLayer extracts a tar file to the destination directory
func extractTarLayer(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip the "." directory entry
		if header.Name == "." {
			continue
		}

		path := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			if err := os.Symlink(header.Linkname, path); err != nil {
				if os.IsExist(err) {
					os.Remove(path)
					if err := os.Symlink(header.Linkname, path); err != nil {
						return err
					}
				} else {
					return err
				}
			}

		default:
			fmt.Printf("Skipping unsupported file type: %c for %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
