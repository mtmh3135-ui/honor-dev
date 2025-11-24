package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/helpers"
	"github.com/mtmh3135/honor/backend/models"
	"golang.org/x/crypto/bcrypt"
)

// ✅ GET all users
func GetUsers(c *fiber.Ctx) error {
	rows, err := config.DB.Query("SELECT user_id, username, role FROM user ORDER BY user_id ASC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil data"})
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.Role); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal baca data"})
		}
		users = append(users, u)
	}

	return c.JSON(fiber.Map{"data": users})
}

// ✅ CREATE user
func CreateUser(c *fiber.Ctx) error {
	var u models.User
	if err := c.BodyParser(&u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Body tidak valid"})
	}

	if u.Username == "" || u.Password == "" || u.Role == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Semua field wajib diisi"})
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	_, err := config.DB.Exec("INSERT INTO user (username, password, role) VALUES (?, ?, ?)", u.Username, hashed, u.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal tambah user"})
	}
	// Ambil userID dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Add User", "Add User "+u.Username); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}
	return c.JSON(fiber.Map{"message": "User berhasil dibuat"})
}

// ✅ UPDATE user
func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var u models.User
	if err := c.BodyParser(&u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Body tidak valid"})
	}

	if u.Username == "" || u.Role == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username dan role wajib diisi"})
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	_, err := config.DB.Exec("UPDATE user SET username = ?, role = ?,password=? WHERE user_id = ?", u.Username, u.Role, hashed, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal update user"})
	}
	// Ambil userID dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Update User", "Update User "+u.Username); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}
	return c.JSON(fiber.Map{"message": "User berhasil diupdate"})
}

// ✅ DELETE user
func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := config.DB.Exec("DELETE FROM user WHERE user_id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal hapus user"})
	}
	// Ambil userID dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Delete User", "Delete User "+id); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}
	return c.JSON(fiber.Map{"message": "User berhasil dihapus"})
}
