package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoOwner = "nokusukun"
	repoName  = "uhh"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
	HTMLURL string  `json:"html_url"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	DownloadURL    string
	AssetName      string
	ReleaseURL     string
	HasUpdate      bool
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
	}

	// Compare versions (strip 'v' prefix if present)
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	if current == "" || current == "dev" {
		// Development version, always offer update
		info.HasUpdate = true
	} else if current != latest && compareVersions(current, latest) < 0 {
		info.HasUpdate = true
	}

	// Find the appropriate asset for this platform
	assetName := getAssetName(release.TagName)
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			info.DownloadURL = asset.BrowserDownloadURL
			info.AssetName = asset.Name
			break
		}
	}

	if info.HasUpdate && info.DownloadURL == "" {
		return nil, fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return info, nil
}

// PerformUpdate downloads and installs the update
func PerformUpdate(info *UpdateInfo) error {
	if !info.HasUpdate {
		return fmt.Errorf("no update available")
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "uhh-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the release
	archivePath := filepath.Join(tempDir, info.AssetName)
	if err := downloadFile(info.DownloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract the binary
	binaryPath := filepath.Join(tempDir, getBinaryName())
	if strings.HasSuffix(info.AssetName, ".zip") {
		if err := extractZip(archivePath, tempDir); err != nil {
			return fmt.Errorf("failed to extract update: %w", err)
		}
	} else if strings.HasSuffix(info.AssetName, ".tar.gz") {
		if err := extractTarGz(archivePath, tempDir); err != nil {
			return fmt.Errorf("failed to extract update: %w", err)
		}
	}

	// Verify the binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("extracted binary not found at %s", binaryPath)
	}

	// Replace the current binary
	if err := replaceBinary(execPath, binaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func fetchLatestRelease() (*Release, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetName(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("uhh-%s-%s-%s%s", version, goos, goarch, ext)
}

func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "uhh.exe"
	}
	return "uhh"
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func replaceBinary(currentPath, newPath string) error {
	// On Windows, we can't replace a running binary directly
	// So we rename the current one first, then move the new one
	if runtime.GOOS == "windows" {
		backupPath := currentPath + ".old"

		// Remove old backup if exists
		os.Remove(backupPath)

		// Rename current to backup
		if err := os.Rename(currentPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup current binary: %w", err)
		}

		// Move new binary to target
		if err := copyFile(newPath, currentPath); err != nil {
			// Try to restore backup
			os.Rename(backupPath, currentPath)
			return fmt.Errorf("failed to install new binary: %w", err)
		}

		// Schedule cleanup of old binary (best effort)
		// The .old file will remain until next update or manual cleanup
		return nil
	}

	// On Unix, we can use atomic rename
	// First copy to temp location in same directory
	dir := filepath.Dir(currentPath)
	tmpPath := filepath.Join(dir, ".uhh-update-tmp")

	if err := copyFile(newPath, tmpPath); err != nil {
		return err
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, currentPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// compareVersions compares two semver strings
// Returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Simple version comparison - split by dots and compare numerically
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}
