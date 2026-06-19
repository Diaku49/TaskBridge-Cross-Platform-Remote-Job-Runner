package executor

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
	"taskbridge/internal/model"
)

type ChecksumExecutor struct{}

type ChecksumPayload struct {
	Path      string `json:"path"`
	Algorithm string `json:"algorithm,omitempty"`
}

func NewChecksumExecutor() *ChecksumExecutor {
	return &ChecksumExecutor{}
}

func (e *ChecksumExecutor) Type() model.JobType {
	return model.JobChecksum
}

func (e *ChecksumExecutor) Execute(ctx context.Context, job model.Job) Result {
	logs := make([]string, 0, 5)

	var payload ChecksumPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   []string{"failed to decode payload: " + err.Error()},
		}
	}

	logs = append(logs, "decoded payload successfully")

	if err := CheckChecksumPayload(&payload); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "invalid payload",
			Logs:   append(logs, "payload validation failed: "+err.Error()),
		}
	}

	hasher := newChecksumHash(payload.Algorithm)
	checksum, bytesRead, err := checksumFile(ctx, payload.Path, hasher)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  "failed to calculate checksum",
			Logs:   append(logs, "failed to calculate checksum: "+err.Error()),
			Result: map[string]any{
				"path":      payload.Path,
				"algorithm": payload.Algorithm,
			},
		}
	}

	logs = append(logs, fmt.Sprintf("calculated %s checksum successfully", payload.Algorithm))

	return Result{
		Status: model.JobSuccess,
		Logs:   logs,
		Result: map[string]any{
			"path":       payload.Path,
			"algorithm":  payload.Algorithm,
			"checksum":   checksum,
			"bytes_read": bytesRead,
		},
	}
}

func CheckChecksumPayload(p *ChecksumPayload) error {
	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	p.Algorithm = strings.ToLower(strings.TrimSpace(p.Algorithm))
	if p.Algorithm == "" {
		p.Algorithm = "sha256"
	}

	switch p.Algorithm {
	case "md5", "sha1", "sha256", "sha512":
		return nil
	default:
		return fmt.Errorf("unsupported algorithm: %s", p.Algorithm)
	}
}

func newChecksumHash(algorithm string) hash.Hash {
	switch algorithm {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha512":
		return sha512.New()
	default:
		return sha256.New()
	}
}

func checksumFile(ctx context.Context, path string, h hash.Hash) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	buf := make([]byte, 32*1024)
	var bytesRead int64
	for {
		if err := ctx.Err(); err != nil {
			return "", bytesRead, err
		}

		n, readErr := file.Read(buf)
		if n > 0 {
			bytesRead += int64(n)
			if _, err := h.Write(buf[:n]); err != nil {
				return "", bytesRead, err
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return "", bytesRead, readErr
		}
	}

	return hex.EncodeToString(h.Sum(nil)), bytesRead, nil
}
