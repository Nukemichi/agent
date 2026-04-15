package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"agent-michi/internal/domain"
	"agent-michi/internal/infrastructure/system"
)

const (
	binaryPath  = "/usr/local/bin/sing-box"
	configPath  = "/etc/sing-box/config.json"
	serviceName = "sing-box"

	// githubAPILatest is the GitHub API endpoint for the latest sing-box release.
	githubAPILatest = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"

	// fallbackVersion is used when the GitHub API is unreachable.
	fallbackVersion = "1.13.8"
)

var fallbackURLPatterns = map[string]string{
	"amd64": "https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-linux-amd64.tar.gz",
	"arm64": "https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-linux-arm64.tar.gz",
	"armv7": "https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-linux-armv7.tar.gz",
}

// DeployUseCase orchestrates the full deploy lifecycle.
type DeployUseCase struct {
	binaryMgr domain.BinaryManager
	svcMgr    domain.ServiceManager
}

// NewDeployUseCase creates a new DeployUseCase.
func NewDeployUseCase(binaryMgr domain.BinaryManager, svcMgr domain.ServiceManager) *DeployUseCase {
	return &DeployUseCase{
		binaryMgr: binaryMgr,
		svcMgr:    svcMgr,
	}
}

// Deploy runs the full deploy sequence.
func (d *DeployUseCase) Deploy(ctx context.Context, req domain.DeployRequest) error {
	url, err := resolveBinaryURL(ctx, req.BinaryURL)
	if err != nil {
		return fmt.Errorf("resolve binary url: %w", err)
	}

	if err := d.binaryMgr.EnsureBinary(ctx, url, binaryPath); err != nil {
		return fmt.Errorf("ensure binary: %w", err)
	}

	if err := os.MkdirAll("/etc/sing-box", 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(configPath, req.Config, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := system.WriteUnitFile(serviceName, binaryPath, configPath); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}

	if err := system.DaemonReload(ctx); err != nil {
		return fmt.Errorf("daemon reload: %w", err)
	}

	if err := d.svcMgr.RestartService(ctx, serviceName); err != nil {
		return fmt.Errorf("restart service: %w", err)
	}

	return nil
}

// resolveBinaryURL returns the download URL to use.
//
// Priority:
//  1. Explicit URL in request (not empty, not "latest") → use as-is.
//  2. Empty or "latest" → fetch the latest matching asset URL from GitHub API.
//  3. GitHub API unreachable or no matching asset → fall back to hardcoded URL pattern.
func resolveBinaryURL(ctx context.Context, requested string) (string, error) {
	normalized := strings.TrimSpace(requested)
	if normalized != "" && !strings.EqualFold(normalized, "latest") {
		return normalized, nil
	}

	arch := goarchToRelease(runtime.GOARCH)

	url, err := fetchLatestAssetURL(ctx, arch)
	if err == nil {
		return url, nil
	}

	// Graceful fallback — don't fail the deploy just because GitHub is slow.
	fallbackURL, fbErr := buildFallbackURL(fallbackVersion, arch)
	if fbErr != nil {
		return "", fmt.Errorf("resolve fallback url: %w", fbErr)
	}

	return fallbackURL, nil
}

// fetchLatestAssetURL queries the GitHub Releases API and returns the matching
// BrowserDownloadURL for the current architecture.
func fetchLatestAssetURL(ctx context.Context, arch string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPILatest, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api: unexpected status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"` // e.g. "v1.13.8"
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode github response: %w", err)
	}
	if release.TagName == "" {
		return "", fmt.Errorf("github api: empty tag_name")
	}

	version := normalizeTag(release.TagName)
	expectedName := fmt.Sprintf("sing-box-%s-linux-%s.tar.gz", version, arch)

	for _, asset := range release.Assets {
		if asset.Name == expectedName && asset.BrowserDownloadURL != "" {
			return asset.BrowserDownloadURL, nil
		}
	}

	fallbackNeedle := fmt.Sprintf("linux-%s.tar.gz", arch)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, fallbackNeedle) && asset.BrowserDownloadURL != "" {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("github api: no matching asset for arch %q", arch)
}

func normalizeTag(tag string) string {
	if strings.HasPrefix(tag, "v") {
		return tag[1:]
	}
	return tag
}

func buildFallbackURL(version, arch string) (string, error) {
	pattern, ok := fallbackURLPatterns[arch]
	if !ok {
		return "", fmt.Errorf("unsupported architecture %q", arch)
	}

	cleanVersion := strings.TrimSpace(normalizeTag(version))
	if cleanVersion == "" {
		return "", fmt.Errorf("empty fallback version")
	}

	return fmt.Sprintf(pattern, cleanVersion, cleanVersion), nil
}

// goarchToRelease maps Go's GOARCH values to the arch strings used in
// sing-box release filenames.
func goarchToRelease(goarch string) string {
	switch goarch {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "arm":
		return "armv7"
	default:
		return goarch
	}
}
