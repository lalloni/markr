package fileutils

import (
	"context"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"

	"github.com/lalloni/markr/logging"
)

var (
	deletes []string
)

func DoDeletes(ctx context.Context) {
	for _, file := range deletes {
		Delete(ctx, file)
	}
}

func AddDelete(ctx context.Context, file string) {
	log := logging.ZapLogger(ctx)
	log.Info("adding delete", zap.String("file", file))
	deletes = append(deletes, file)
}

func Delete(ctx context.Context, file string) error {
	log := logging.ZapLogger(ctx)
	log.Info("deleting", zap.String("file", file))
	return os.Remove(file)
}

func NewTempFileName(base, kind, id, ext string) string {
	return path.Join(os.TempDir(), base+"-"+kind+"-"+id+"."+ext)
}

func NewTempFile(base, kind, id, ext string) (*os.File, error) {
	return os.Create(NewTempFileName(base, kind, id, ext))
}

func ChangeExtension(file string, ext string) string {
	return strings.TrimSuffix(file, path.Ext(file)) + "." + ext
}
