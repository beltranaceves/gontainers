package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type Manifest struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		Digest string `json:"digest"`
	} `json:"layers"`
}

func main() {
	repo, tag := "library/alpine", "latest" // Example: Official Alpine image

	token, err := getToken(repo)
	if err != nil {
		panic(err)
	}

	manifest, err := getManifest(repo, tag, token)
	if err != nil {
		panic(err)
	}

	// Download config (image metadata)
	if err := downloadBlob(repo, manifest.Config.Digest, token, "config.json"); err != nil {
		panic(err)
	}

	// Download layers
	for i, layer := range manifest.Layers {
		output := fmt.Sprintf("layer%d.tar.gz", i)
		if err := downloadBlob(repo, layer.Digest, token, output); err != nil {
			panic(err)
		}
	}
}

// Retrieve authentication token
func getToken(repo string) (string, error) {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)
	resp, err := http.Get(authURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

// Fetch image manifest
func getManifest(repo, tag, token string) (*Manifest, error) {
	url := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", repo, tag)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest error: %s", resp.Status)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// Download a blob (layer or config)
func downloadBlob(repo, digest, token, output string) error {
	url := fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s", repo, digest)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blob download failed: %s", resp.Status)
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
