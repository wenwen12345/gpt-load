package keypool

import (
	"gpt-load/internal/channel"
	"gpt-load/internal/encryption"
	"gpt-load/internal/models"
	"gpt-load/internal/store"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestApplyValidationResultPersistsOpenAITierWithoutCacheEntry(t *testing.T) {
	db := newProviderTestDB(t)
	encryptionSvc, err := encryption.NewService("")
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}

	apiKey := models.APIKey{
		KeyValue: "sk-test",
		KeyHash:  encryptionSvc.Hash("sk-test"),
		GroupID:  1,
		Status:   models.KeyStatusActive,
	}
	if err := db.Create(&apiKey).Error; err != nil {
		t.Fatalf("failed to create api key: %v", err)
	}

	provider := NewProvider(db, store.NewMemoryStore(), nil, encryptionSvc)
	result := channel.KeyValidationResult{
		IsValid:           true,
		OpenAITierUpdated: true,
		OpenAITier:        "T3",
	}

	if err := provider.ApplyValidationResult(&apiKey, &models.Group{ID: 1}, result, ""); err != nil {
		t.Fatalf("failed to apply validation result: %v", err)
	}

	var persisted models.APIKey
	if err := db.First(&persisted, apiKey.ID).Error; err != nil {
		t.Fatalf("failed to load persisted api key: %v", err)
	}
	if persisted.OpenAITier != "T3" {
		t.Fatalf("unexpected persisted OpenAI tier: got %q want %q", persisted.OpenAITier, "T3")
	}

	cached, err := provider.store.HGetAll("key:1")
	if err != nil {
		t.Fatalf("failed to load cached api key: %v", err)
	}
	if cached["openai_tier"] != "T3" {
		t.Fatalf("unexpected cached OpenAI tier: got %q want %q", cached["openai_tier"], "T3")
	}
}

func TestApplyValidationResultKeepsExistingOpenAITierWhenInferenceIsEmpty(t *testing.T) {
	db := newProviderTestDB(t)
	encryptionSvc, err := encryption.NewService("")
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}

	apiKey := models.APIKey{
		KeyValue:   "sk-test",
		KeyHash:    encryptionSvc.Hash("sk-test"),
		GroupID:    1,
		Status:     models.KeyStatusActive,
		OpenAITier: "T2",
	}
	if err := db.Create(&apiKey).Error; err != nil {
		t.Fatalf("failed to create api key: %v", err)
	}

	provider := NewProvider(db, store.NewMemoryStore(), nil, encryptionSvc)
	result := channel.KeyValidationResult{
		IsValid:           true,
		OpenAITierUpdated: true,
		OpenAITier:        "",
	}

	if err := provider.ApplyValidationResult(&apiKey, &models.Group{ID: 1}, result, ""); err != nil {
		t.Fatalf("failed to apply validation result: %v", err)
	}

	var persisted models.APIKey
	if err := db.First(&persisted, apiKey.ID).Error; err != nil {
		t.Fatalf("failed to load persisted api key: %v", err)
	}
	if persisted.OpenAITier != "T2" {
		t.Fatalf("unexpected persisted OpenAI tier: got %q want %q", persisted.OpenAITier, "T2")
	}
}

func newProviderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.APIKey{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}
