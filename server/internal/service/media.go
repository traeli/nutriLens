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

	model "shijibu/internal/model/pgsql"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	item := &model.RecordMedia{RecordID: recordID, RecordVersionID: view.Version.ID, ObjectKey: objectKey, PublicURL: publicURL, MediaType: mediaType, SafetyStatus: "pending", DesensitizeStatus: "pending", CreatedAt: time.Now()}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var record model.VisitRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		for _, media := range record.Media {
			if media.ID >= item.ID {
				item.ID = media.ID + 1
			}
		}
		if item.ID == 0 {
			item.ID = 1
		}
		record.Media = append(record.Media, *item)
		return tx.Model(&record).Update("media", record.Media).Error
	}); err != nil {
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
	item := &model.RecordEvidence{RecordID: recordID, UserID: userID, EvidenceType: evidenceType, OriginalObjectKey: objectKey, VerifyStatus: "pending", RetentionUntil: &retention, CreatedAt: time.Now()}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var record model.VisitRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		for _, evidence := range record.Evidences {
			if evidence.ID >= item.ID {
				item.ID = evidence.ID + 1
			}
		}
		if item.ID == 0 {
			item.ID = 1
		}
		record.Evidences = append(record.Evidences, *item)
		return tx.Model(&record).Update("evidences", record.Evidences).Error
	}); err != nil {
		_ = os.Remove(filepath.Join(s.root, filepath.FromSlash(objectKey)))
		return nil, err
	}
	return item, nil
}

func (s *MediaService) DeleteRecordMedia(userID, recordID, mediaID uint) error {
	var item model.RecordMedia
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var record model.VisitRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		media := make(model.JSONList[model.RecordMedia], 0, len(record.Media))
		found := false
		for _, candidate := range record.Media {
			if candidate.ID == mediaID {
				item = candidate
				found = true
				continue
			}
			media = append(media, candidate)
		}
		if !found {
			return ErrNotFound
		}
		return tx.Model(&record).Update("media", media).Error
	})
	if err != nil {
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
