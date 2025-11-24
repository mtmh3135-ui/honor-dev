package helpers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/models"
)

func LogActivity(c *fiber.Ctx, userID int64, action, description string) error {

	query := `
		INSERT INTO user_activity (user_id, action, description, created_at)
		VALUES (?, ?, ?, NOW())
	`
	result, err := config.DB.Exec(query, userID, action, description)
	if err != nil {
		fmt.Println("Failed to insert log:", err)
	}
	rows, _ := result.RowsAffected()
	fmt.Println("✅ Activity inserted:", rows, "rows")

	return err
}

// GET /api/activity
func GetUserActivity(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	rows, err := config.DB.Query(`
		SELECT  user_id, action, description, , created_at
		FROM user_activity_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var log models.Activity
		rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Description, &log.CreatedAt)
		activities = append(activities, log)
	}

	return c.JSON(activities)
}
