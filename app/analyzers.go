package app

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/mostlygeek/go-exiftool"
	"github.com/vansante/go-ffprobe"
)

func DigestMD5(_ context.Context, file *os.File) (res string, err error) {
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	hash := md5.New()

	if _, err = io.Copy(hash, file); err != nil {
		return
	}

	hashInBytes := hash.Sum(nil)[:16]

	res = hex.EncodeToString(hashInBytes)

	return
}

func DigestSHA1(_ context.Context, file *os.File) (res string, err error) {
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	hash := sha1.New()

	if _, err = io.Copy(hash, file); err != nil {
		return
	}

	hashInBytes := hash.Sum(nil)[:16]

	res = hex.EncodeToString(hashInBytes)

	return
}

func Mimetype(_ context.Context, file *os.File) (string, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	return mime.String(), nil
}

func EXIF(_ context.Context, file *os.File, et *exiftool.Stayopen) (map[string]interface{}, error) {
	etResults, err := et.Extract(file.Name())
	if err != nil {
		return nil, fmt.Errorf("exif extract failed: %w", err)
	}
	attributes := make([]map[string]interface{}, 0)
	err = json.Unmarshal(etResults, &attributes)
	if err != nil {
		return nil, fmt.Errorf("exif decode failed: %w", err)
	}
	if len(attributes) == 0 {
		return nil, nil
	}

	etResult := attributes[0]

	delete(etResult, "FileName")
	delete(etResult, "SourceFile")
	delete(etResult, "Directory")
	delete(etResult, "FilePermissions")
	delete(etResult, "FileAccessDate")
	delete(etResult, "FileInodeChangeDate")
	delete(etResult, "FileModifyDate")

	for k, v := range etResult {
		if vStr, ok := v.(string); ok {
			if strings.Contains(vStr, "use -b option to extract") {
				etResult[k] = "[binary data not parsed]"
			}
		}
	}

	return etResult, nil
}

func FFProbe(ctx context.Context, file *os.File) (*ffprobe.ProbeData, error) {
	info, err := ffprobe.GetProbeDataContext(ctx, file.Name())
	if err != nil {
		return nil, err
	}

	if info.Format != nil {
		parts := strings.Split(info.Format.Filename, "/")
		info.Format.Filename = parts[len(parts)-1]
	}

	return info, nil
}
