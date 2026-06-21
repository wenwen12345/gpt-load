package db

import (
	"fmt"
	"gpt-load/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// V1_1_3_TierColumnLength widens tier for balance labels such as "¥12.34 / $5".
func V1_1_3_TierColumnLength(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&models.APIKey{}, "tier") {
		return nil
	}

	if err := db.Migrator().AlterColumn(&models.APIKey{}, "Tier"); err != nil {
		return fmt.Errorf("failed to alter tier column: %w", err)
	}

	logrus.Info("Migration v1.1.3 completed successfully")
	return nil
}
