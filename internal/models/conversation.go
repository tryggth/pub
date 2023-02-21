package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// A Conversation is a collection of related statuses. It is a way to group
// together statuses that are replies to each other, or that are part of the
// same thread of conversation. Conversations are not necessarily public, and
// may be limited to a set of participants.
type Conversation struct {
	ID         uint32 `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Visibility ConversationVisibility `gorm:"not null"`
}

type ConversationVisibility string

func (ConversationVisibility) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "postgres":
		return "enum('public', 'unlisted', 'private', 'direct', 'limited')"
	case "sqlite":
		return "TEXT"
	default:
		return ""
	}
}

type Conversations struct {
	db *gorm.DB
}

func NewConversations(db *gorm.DB) *Conversations {
	return &Conversations{
		db: db,
	}
}

// New returns a new Conversations with the given visibility.
func (c *Conversations) New(vis string) (*Conversation, error) {
	conv := Conversation{
		Visibility: ConversationVisibility(vis),
	}
	if err := c.db.Create(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

func (c *Conversations) FindOrCreate(id uint32, vis string) (*Conversation, error) {
	var conversation Conversation
	if err := c.db.FirstOrCreate(&conversation, Conversation{
		Visibility: ConversationVisibility(vis),
	}).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}
