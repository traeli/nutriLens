package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"shijibu/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaService struct {
	db         *gorm.DB
	root       string
	publicBase string
}

func NewMediaService(db *gorm.DB, root, publicBase string) *MediaService {
	return &MediaService{db: db, root: root, publicBase: strings.TrimRight(publicBase, "/")}
}

func (s *MediaService) AddRecordMedia(userID, recordID uint, header *multipart.FileHeader, mediaType string) (*model.RecordMedia, error) {
	if header == nil || header.Size < 1 || header.Size > 10<<20 || (mediaType != "photo" && mediaType != "receipt") {
		return nil, ErrInvalidInput
	}
	view, err := NewRecordService(s.db).GetOwned(userID, recordID)
	if err != nil {
		return nil, err
	}
	if view.Record.PublishStatus != "draft" && view.Record.PublishStatus != "rejected" {
		return nil, ErrConflict
	}
	objectKey, publicURL, err := s.save(header, "public/records")
	if err != nil {
		return nil, err
	}
	item := &model.RecordMedia{RecordID: recordID, RecordVersionID: view.Version.ID, ObjectKey: objectKey, PublicURL: publicURL, MediaType: mediaType, SafetyStatus: "pending", DesensitizeStatus: "pending"}
	if err := s.db.Create(item).Error; err != nil {
		_ = os.Remove(filepath.Join(s.root, filepath.FromSlash(objectKey)))
		return nil, err
	}
	return item, nil
}

func (s *MediaService) AddEvidence(userID, recordID uint, header *multipart.FileHeader, evidenceType string) (*model.RecordEvidence, error) {
	if header == nil || header.Size < 1 || header.Size > 10<<20 || (evidenceType != "receipt" && evidenceType != "order" && evidenceType != "other") {
		return nil, ErrInvalidInput
	}
	if _, err := NewRecordService(s.db).GetOwned(userID, recordID); err != nil {
		return nil, err
	}
	objectKey, _, err := s.save(header, "private/evidence")
	if err != nil {
		return nil, err
	}
	retention := time.Now().AddDate(0, 6, 0)
	item := &model.RecordEvidence{RecordID: recordID, UserID: userID, EvidenceType: evidenceType, OriginalObjectKey: objectKey, VerifyStatus: "pending", RetentionUntil: &retention}
	if err := s.db.Create(item).Error; err != nil {
		_ = os.Remove(filepath.Join(s.root, filepath.FromSlash(objectKey)))
		return nil, err
	}
	return item, nil
}

func (s *MediaService) DeleteRecordMedia(userID, recordID, mediaID uint) error {
	if _, err := NewRecordService(s.db).GetOwned(userID, recordID); err != nil {
		return err
	}
	var item model.RecordMedia
	if err := s.db.Where("id = ? AND record_id = ?", mediaID, recordID).First(&item).Error; err != nil {
		return mapNotFound(err)
	}
	if err := s.db.Delete(&item).Error; err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(s.root, filepath.FromSlash(item.ObjectKey))); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *MediaService) save(header *multipart.FileHeader, prefix string) (string, string, error) {
	source, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer source.Close()
	probe := make([]byte, 512)
	n, _ := io.ReadFull(source, probe)
	contentType := http.DetectContentType(probe[:n])
	allowed := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	ext, ok := allowed[contentType]
	if !ok {
		return "", "", ErrInvalidInput
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}
	objectKey := filepath.ToSlash(filepath.Join(prefix, time.Now().Format("2006/01"), uuid.NewString()+ext))
	targetPath := filepath.Join(s.root, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
		return "", "", err
	}
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", "", err
	}
	defer target.Close()
	if _, err := io.Copy(target, io.LimitReader(source, 10<<20)); err != nil {
		return "", "", fmt.Errorf("store upload: %w", err)
	}
	publicKey := strings.TrimPrefix(objectKey, "public/")
	return objectKey, s.publicBase + "/" + publicKey, nil
}
