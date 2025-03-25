package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ImageReference represents a Docker image reference (e.g., "ubuntu:latest")
type ImageReference struct {
	Registry string
	Repo     string
	Tag      string
}

// Manifest represents a Docker image manifest
type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

// ParseImageReference parses an image reference string into its components
func ParseImageReference(ref string) ImageReference {
	registry := "registry-1.docker.io"
	repo := ref
	tag := "latest"

	// Handle tag
	if parts := strings.Split(ref, ":"); len(parts) > 1 {
		repo = parts[0]
		tag = parts[1]
	}

	// Handle registry
	if parts := strings.Split(repo, "/"); len(parts) > 1 && strings.Contains(parts[0], ".") {
		registry = parts[0]
		repo = strings.Join(parts[1:], "/")
	}

	// Add library/ prefix for official images
	if !strings.Contains(repo, "/") {
		repo = "library/" + repo
	}

	return ImageReference{
		Registry: registry,
		Repo:     repo,
		Tag:      tag,
	}
}

// DownloadImage downloads a Docker image from Docker Hub and extracts its layers
func DownloadImage(imageRef string, destDir string) error {
	ref := ParseImageReference(imageRef)

	fmt.Printf("Downloading image %s:%s...\n", ref.Repo, ref.Tag)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Create a temporary directory for downloaded layers
	tempDir := filepath.Join(destDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory when done

	// Get auth token
	fmt.Println("Getting authentication token...")
	token, err := getAuthToken(ref)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %v", err)
	}

	// Get manifest
	fmt.Println("Fetching image manifest...")
	manifest, err := getManifest(ref, token)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %v", err)
	}

	// Create rootfs directory
	rootfsDir := filepath.Join(destDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return fmt.Errorf("failed to create rootfs directory: %v", err)
	}

	// Download and extract layers
	fmt.Println("Downloading and extracting layers...")
	for i, layer := range manifest.Layers {
		layerPath := filepath.Join(tempDir, fmt.Sprintf("layer_%d.tar", i))

		fmt.Printf("Downloading layer %d of %d: %s\n", i+1, len(manifest.Layers), layer.Digest)
		if err := downloadBlob(ref, layer.Digest, token, layerPath); err != nil {
			return fmt.Errorf("failed to download layer %s: %v", layer.Digest, err)
		}

		fmt.Printf("Extracting layer %d to rootfs...\n", i+1)
		if err := extractLayer(layerPath, rootfsDir); err != nil {
			return fmt.Errorf("failed to extract layer %s: %v", layer.Digest, err)
		}
	}

	// Download image config
	fmt.Println("Downloading image configuration...")
	configPath := filepath.Join(destDir, "config.json")
	if err := downloadBlob(ref, manifest.Config.Digest, token, configPath); err != nil {
		return fmt.Errorf("failed to download config: %v", err)
	}

	// Save manifest
	manifestPath := filepath.Join(destDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %v", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %v", err)
	}

	// Create image info file
	imageInfo := map[string]string{
		"name": ref.Repo,
		"tag":  ref.Tag,
	}
	infoData, err := json.MarshalIndent(imageInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal image info: %v", err)
	}

	infoPath := filepath.Join(destDir, "image.json")
	if err := os.WriteFile(infoPath, infoData, 0644); err != nil {
		return fmt.Errorf("failed to write image info: %v", err)
	}

	fmt.Printf("Successfully downloaded and extracted image %s:%s to %s\n", ref.Repo, ref.Tag, destDir)
	return nil
}

// getAuthToken gets an authentication token for Docker Hub
func getAuthToken(ref ImageReference) (string, error) {
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", ref.Repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

// getManifest gets the manifest for a Docker image
func getManifest(ref ImageReference, token string) (*Manifest, error) {
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", ref.Registry, ref.Repo, ref.Tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("manifest request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// downloadBlob downloads a blob (layer or config) from a Docker registry
func downloadBlob(ref ImageReference, digest string, token string, destPath string) error {
	url := fmt.Sprintf("https://%s/v2/%s/blobs/%s", ref.Registry, ref.Repo, digest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If we get a 307 redirect, follow it manually
		if resp.StatusCode == http.StatusTemporaryRedirect {
			redirectURL := resp.Header.Get("Location")
			if redirectURL != "" {
				fmt.Printf("Following redirect to: %s\n", redirectURL)
				redirectResp, err := http.Get(redirectURL)
				if err != nil {
					return err
				}
				defer redirectResp.Body.Close()

				if redirectResp.StatusCode != http.StatusOK {
					bodyBytes, _ := io.ReadAll(redirectResp.Body)
					return fmt.Errorf("blob download failed with status: %s, body: %s", redirectResp.Status, string(bodyBytes))
				}

				out, err := os.Create(destPath)
				if err != nil {
					return err
				}
				defer out.Close()

				_, err = io.Copy(out, redirectResp.Body)
				return err
			}
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("blob download failed with status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractLayer extracts a layer tarball to the rootfs directory
func extractLayer(layerPath, rootfsDir string) error {
	file, err := os.Open(layerPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Docker layers are typically gzipped tar files
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		// If not gzipped, try as a regular tar
		file.Seek(0, 0)
		return extractTar(file, rootfsDir)
	}
	defer gzipReader.Close()

	return extractTar(gzipReader, rootfsDir)
}

// extractTar extracts a tar archive to the specified directory
func extractTar(reader io.Reader, destDir string) error {
	tarReader := tar.NewReader(reader)

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

		// Handle whiteout files (used by Docker to remove files)
		if strings.HasPrefix(filepath.Base(header.Name), ".wh.") {
			// Get the file/dir to be removed
			nameToRemove := strings.Replace(filepath.Base(header.Name), ".wh.", "", 1)
			// Get the path to the file/dir to be removed
			pathToRemove := filepath.Join(destDir, filepath.Dir(header.Name), nameToRemove)

			// Remove the file/dir
			if err := os.RemoveAll(pathToRemove); err != nil {
				// Don't fail if the file doesn't exist
				if !os.IsNotExist(err) {
					return err
				}
			}
			continue
		}

		// Construct the path for the file/directory
		path := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			// Create file
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// Copy file contents
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()

		case tar.TypeSymlink:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			// Create symlink
			if err := os.Symlink(header.Linkname, path); err != nil {
				// If the symlink already exists, remove it and try again
				if os.IsExist(err) {
					os.Remove(path)
					if err := os.Symlink(header.Linkname, path); err != nil {
						return err
					}
				} else {
					return err
				}
			}

		case tar.TypeLink:
			// Create hard link
			linkTarget := filepath.Join(destDir, header.Linkname)
			if err := os.Link(linkTarget, path); err != nil {
				// If the link already exists, remove it and try again
				if os.IsExist(err) {
					os.Remove(path)
					if err := os.Link(linkTarget, path); err != nil {
						// If it still fails, just copy the file
						if srcFile, err := os.Open(linkTarget); err == nil {
							defer srcFile.Close()
							if destFile, err := os.Create(path); err == nil {
								defer destFile.Close()
								io.Copy(destFile, srcFile)
							}
						}
					}
				}
			}

		default:
			fmt.Printf("Skipping unsupported file type: %c for %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

func main() {
	// Download Ubuntu latest image to ./images/ubuntu directory
	err := DownloadImage("alpine:latest", "./images/alpine")
	if err != nil {
		fmt.Printf("Error downloading image: %v\n", err)
	}
}
