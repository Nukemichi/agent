package system

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// BinaryManager implements domain.BinaryManager.
type BinaryManager struct{}

// NewBinaryManager creates a new BinaryManager.
func NewBinaryManager() *BinaryManager {
	return &BinaryManager{}
}

// EnsureBinary downloads and extracts the binary from url to targetPath if not already present.
// The url must point to a .tar.gz archive produced by the sing-box release pipeline.
func (b *BinaryManager) EnsureBinary(ctx context.Context, url, targetPath string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download archive: unexpected status %d", resp.StatusCode)
	}

	// Stream-decompress: response body → gzip → tar → target binary
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		// The sing-box archive contains a single directory; the binary lives at
		// <dir>/sing-box. Match any entry whose base name is exactly "sing-box".
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != filepath.Base(targetPath) {
			continue
		}

		if err := writeAtomically(tr, targetPath); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("binary %q not found inside archive", filepath.Base(targetPath))
}

// writeAtomically writes src to a temp file, chmods it, then renames to dst.
func writeAtomically(src io.Reader, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), "agent-binary-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName) // no-op after successful rename
	}()

	if _, err := io.Copy(tmp, src); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpName, 0755); err != nil {
		return fmt.Errorf("chmod binary: %w", err)
	}

	if err := os.Rename(tmpName, dst); err != nil {
		return fmt.Errorf("move binary to target: %w", err)
	}

	return nil
}
