package articlerepository

import (
	"article-versioning-api/core/entity"
	"time"
)

type Version struct {
	Serial               string     `json:"serial"`
	AuthorUsername       string     `json:"authorUsername"`
	VersionNumber        int        `json:"versionNumber"`
	ArticleSerial        string     `json:"articleSerial"`
	Title                string     `json:"title"`
	Content              string     `json:"content"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            *time.Time `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt"`
	PublishedAt          *time.Time `json:"publishedAt"`
	TagRelationshipScore float32    `json:"tagRelationshipScore"`
}

func (v *Version) parseToVersion() *entity.Version {
	return &entity.Version{
		Serial:               v.Serial,
		AuthorUsername:       v.AuthorUsername,
		VersionNumber:        v.VersionNumber,
		ArticleSerial:        v.ArticleSerial,
		Title:                v.Title,
		Content:              v.Content,
		Status:               v.Status,
		CreatedAt:            v.CreatedAt,
		UpdatedAt:            v.UpdatedAt,
		DeletedAt:            v.DeletedAt,
		PublishedAt:          v.PublishedAt,
		TagRelationshipScore: v.TagRelationshipScore,
	}
}

type VersionTag struct {
	VersionSerial string
	TagSerial     string
	TagName       string
}
