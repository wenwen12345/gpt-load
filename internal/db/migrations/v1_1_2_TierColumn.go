package db

import (
	"fmt"
	"gpt-load/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// V1_1_2_TierColumn backfills the generic tier column from OpenAI-specific data.
func V1_1_2_TierColumn(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&models.APIKey{}, "tier") {
		if err := db.Migrator().AddColumn(&models.APIKey{}, "Tier"); err != nil {
			return fmt.Errorf("failed to add tier column: %w", err)
		}
	}
	if !db.Migrator().HasColumn(&models.APIKey{}, "openai_tier") {
		return nil
	}

	if err := db.Exec(`
		UPDATE api_keys
		SET tier = openai_tier
		WHERE (tier IS NULL OR tier = '')
		  AND openai_tier IS NOT NULL
		  AND openai_tier != ''
	`).Error; err != nil {
		return fmt.Errorf("failed to migrate tier values: %w", err)
	}

	logrus.Info("Migration v1.1.2 completed successfully")
	return nil
}
