package db

import (
	"fmt"
	"gpt-load/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// V1_1_1_OpenAITierColumn copies values from the accidental open_ai_tier column.
func V1_1_1_OpenAITierColumn(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&models.APIKey{}, "open_ai_tier") {
		return nil
	}
	if !db.Migrator().HasColumn(&models.APIKey{}, "openai_tier") {
		if err := db.Migrator().AddColumn(&models.APIKey{}, "OpenAITier"); err != nil {
			return fmt.Errorf("failed to add openai_tier column: %w", err)
		}
	}

	if err := db.Exec(`
		UPDATE api_keys
		SET openai_tier = open_ai_tier
		WHERE (openai_tier IS NULL OR openai_tier = '')
		  AND open_ai_tier IS NOT NULL
		  AND open_ai_tier != ''
	`).Error; err != nil {
		return fmt.Errorf("failed to migrate OpenAI tier values: %w", err)
	}

	logrus.Info("Migration v1.1.1 completed successfully")
	return nil
}
