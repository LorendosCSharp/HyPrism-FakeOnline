package util

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ENTITLEMENTS can be reused from server/app
var ENTITLEMENTS = []string{"game.base", "game.deluxe", "game.founder"}

// ProfileInfo represents the identity payload
type ProfileInfo struct {
	Username     string   `json:"username"`
	Entitlements []string `json:"entitlements"`
	Skin         string   `json:"skin"`
}

// JWT header struct
type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// Identity token struct
type identityToken struct {
	Exp     int         `json:"exp"`
	Iat     int         `json:"iat"`
	Iss     string      `json:"iss"`
	Jti     string      `json:"jti"`
	Scope   string      `json:"scope"`
	Sub     string      `json:"sub"`
	Profile ProfileInfo `json:"profile"`
}

// Session token struct
type sessionToken struct {
	Exp   int    `json:"exp"`
	Iat   int    `json:"iat"`
	Iss   string `json:"iss"`
	Jti   string `json:"jti"`
	Scope string `json:"scope"`
	Sub   string `json:"sub"`
}

// b64json encodes a struct to base64 JSON
func b64json(v interface{}) string {
	b, _ := json.Marshal(v)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// fakeSign generates a random signature string
func fakeSign() string {
	data := make([]byte, 0x40)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(data)
}

// UsernameToUuid generates a reproducible UUID from a username
func UsernameToUuid(username string) string {
	m := md5.New()
	m.Write([]byte(username))
	h := hex.EncodeToString(m.Sum(nil))
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// GenerateIdentityJwt creates a fake identity JWT
func GenerateIdentityJwt(scope, username, skin string) string {
	head := jwtHeader{
		Alg: "EdDSA",
		Kid: "2025-10-01",
		Typ: "JWT",
	}

	idTok := identityToken{
		Exp:   int(time.Now().Add(time.Hour * 10).Unix()),
		Iat:   int(time.Now().Unix()),
		Iss:   "https://sessions.hytale.com",
		Jti:   uuid.NewString(),
		Scope: scope,
		Sub:   UsernameToUuid(username),
		Profile: ProfileInfo{
			Username:     username,
			Entitlements: ENTITLEMENTS,
			Skin:         skin,
		},
	}

	return b64json(head) + "." + b64json(idTok) + "." + fakeSign()
}

// GenerateSessionJwt creates a fake session JWT
func GenerateSessionJwt(scope, username string) string {
	head := jwtHeader{
		Alg: "EdDSA",
		Kid: "2025-10-01",
		Typ: "JWT",
	}

	sesTok := sessionToken{
		Exp:   int(time.Now().Add(time.Hour * 10).Unix()),
		Iat:   int(time.Now().Unix()),
		Iss:   "https://sessions.hytale.com",
		Jti:   uuid.NewString(),
		Scope: scope,
		Sub:   UsernameToUuid(username),
	}

	return b64json(head) + "." + b64json(sesTok) + "." + fakeSign()
}

// VerifySHA256 verifies the SHA256 checksum of a file
func VerifySHA256(filePath, expectedHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	expectedHash = strings.ToLower(strings.TrimSpace(expectedHash))

	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// ExtractZip extracts a ZIP archive to a destination directory
func ExtractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		// Security check for path traversal
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
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

// ExtractTarGz extracts a tar.gz archive to a destination directory
func ExtractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)

		// Security check for path traversal
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			os.Symlink(header.Linkname, path)
		}
	}

	return nil
}

// ExtractArchive extracts an archive based on its extension
func ExtractArchive(src, dest string) error {
	if strings.HasSuffix(src, ".zip") {
		return ExtractZip(src, dest)
	} else if strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
		return ExtractTarGz(src, dest)
	}
	return fmt.Errorf("unsupported archive format: %s", src)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Preserve permissions
	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, info.Mode())
	}

	return nil
}

// CopyDir copies a directory recursively
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// HideConsoleWindow hides the console window on Windows
// Implementation is in util_windows.go and util_unix.go
