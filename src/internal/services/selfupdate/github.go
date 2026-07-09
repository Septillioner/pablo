package selfupdate

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseResponse struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAssets struct {
	Tag         string
	AssetName   string
	DownloadURL string
	ChecksumURL string
}

type githubClient struct {
	httpClient *http.Client
}

func newGitHubClient() *githubClient {
	return &githubClient{httpClient: &http.Client{}}
}

func (c *githubClient) fetchRelease(tag string) (*releaseResponse, error) {
	url := githubAPIBase + "/latest"
	if tag != "" {
		url = githubAPIBase + "/tags/" + releaseTag(tag)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", githubUserAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("github api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode github release: %w", err)
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("release response missing tag_name")
	}

	return &release, nil
}

func (c *githubClient) resolveAssets(tag string) (*releaseAssets, error) {
	assetName, err := platformAssetName()
	if err != nil {
		return nil, err
	}

	release, err := c.fetchRelease(tag)
	if err != nil {
		return nil, err
	}

	result := &releaseAssets{
		Tag:       release.TagName,
		AssetName: assetName,
	}

	for _, asset := range release.Assets {
		switch asset.Name {
		case assetName:
			result.DownloadURL = asset.BrowserDownloadURL
		case checksumsFileName:
			result.ChecksumURL = asset.BrowserDownloadURL
		}
	}

	if result.DownloadURL == "" {
		return nil, fmt.Errorf("release %s has no asset %s", release.TagName, assetName)
	}

	return result, nil
}

func (c *githubClient) download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", githubUserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download: %w", err)
	}
	return data, nil
}

func parseExpectedChecksum(checksums []byte, assetName string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(checksums))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		if parts[1] == assetName {
			return strings.ToLower(parts[0]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("checksum for %s not found in checksums.txt", assetName)
}

func verifyChecksum(data []byte, expected string) error {
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if actual != strings.ToLower(expected) {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

func (c *githubClient) downloadVerifiedBinary(assets *releaseAssets) (string, error) {
	data, err := c.download(assets.DownloadURL)
	if err != nil {
		return "", err
	}

	if assets.ChecksumURL != "" {
		checksums, err := c.download(assets.ChecksumURL)
		if err != nil {
			return "", err
		}
		expected, err := parseExpectedChecksum(checksums, assets.AssetName)
		if err != nil {
			return "", err
		}
		if err := verifyChecksum(data, expected); err != nil {
			return "", fmt.Errorf("%w for %s", err, assets.AssetName)
		}
	}

	tempFile, err := os.CreateTemp("", "pablo-update-*")
	if err != nil {
		return "", err
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return "", err
	}
	if err := os.Chmod(tempPath, 0o755); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	return tempPath, nil
}
