package middleware

import (
	"database/sql"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/models"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func Login(c *fiber.Ctx) error {
	// hashed, _ := bcrypt.GenerateFromPassword([]byte("Admin123"), 12)
	// fmt.Println(string(hashed))
	var body LoginRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user models.User

	// Query ke DB pakai QueryRow (tanpa GORM)
	err := config.DB.QueryRow(
		"SELECT user_id, username, password, role FROM user WHERE username = ? LIMIT 1",
		body.Username,
	).Scan(&user.UserID, &user.Username, &user.Password, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("error: User not found")
			return c.Status(401).JSON(fiber.Map{"error": "User not found"})
		}
		log.Printf("error: Database error")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Cek password — dengan hash di db
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		log.Printf("error: Password salah")
		return c.Status(401).JSON(fiber.Map{"error": "Password salah"})
	}

	//generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.UserID,
		"username": body.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("SECRET_KEY"))

	//  Simpan token sebagai HttpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour),
		HTTPOnly: true,
		Secure:   false, // set true di production (HTTPS)
		SameSite: "Lax",
	})

	// log.Println("user id   : ", user.UserID)
	// log.Println("user name : ", user.Username)

	// Kalau sukses, bisa lanjut generate token / simpan session
	return c.JSON(fiber.Map{
		"message":  "Login success",
		"userid":   user.UserID,
		"username": user.Username,
		"token":    tokenString,
	})
}

func Register(c *fiber.Ctx) error {
	db := c.Locals("db").(*sql.DB)
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Hash password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

	_, err := db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", req.Username, hashed)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan user"})
	}

	return c.JSON(fiber.Map{"message": "User berhasil dibuat"})
}

func AuthRequired(c *fiber.Ctx) error {
	tokenStr := c.Cookies("auth_token")
	if tokenStr == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("SECRET_KEY"), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("user_id", int64(claims["user_id"].(float64)))
	c.Locals("role", claims["role"].(string))
	return c.Next()
}

func Me(c *fiber.Ctx) error {
	tokenStr := c.Cookies("auth_token")
	if tokenStr == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("SECRET_KEY"), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}

	claims := token.Claims.(jwt.MapClaims)
	return c.JSON(fiber.Map{
		"user_id":  claims["user_id"],
		"username": claims["username"],
		"role":     claims["role"],
	})
}

func AdminOnly(c *fiber.Ctx) error {
	cookie := c.Cookies("auth_token")
	if cookie == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Token tidak ditemukan"})
	}

	token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		return []byte("SECRET_KEY"), nil
	})
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Token tidak valid"})
	}

	claims := token.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Akses ditolak, bukan admin"})
	}

	// Opsional: simpan info ke context
	c.Locals("username", claims["username"])
	c.Locals("role", claims["role"])
	c.Locals("userid", claims["userid"])

	return c.Next()
}
