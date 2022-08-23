package main

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func setupSql(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY, username TEXT, password TEXT)")
	if err != nil {
		panic(err)
	}
}

func main() {
	app := fiber.New()

	/*
		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			panic(err)
		}

		stmt, err := db.Prepare("INSERT INTO users(username, password) values(?,?)")
		if err != nil {
			panic(err)
		}

		res, err := stmt.Exec("jack", "pass")
		if err != nil {
			panic(err)
		}

		fmt.Println(res.LastInsertId())
	*/
	storage := sqlite3.New(sqlite3.Config{
		Database: "./database.db",
		Table:    "sessions",
	})
	store := session.New(session.Config{
		Storage: storage,
	})

	db := storage.Conn()

	setupSql(db)

	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			panic(err)
		}

		username := sess.Get("username")

		if username != nil {
			return c.SendString(fmt.Sprintf("Hello, %s", username.(string)))
		}
		return c.SendString("You are a stranger")
	})

	app.Post("/signup", func(c *fiber.Ctx) error {

		// Parse body and sanitize inputs
		payload := User{}
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if payload.Username == "" || payload.Password == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// Check for duplicate username
		stmt, err := db.Prepare("SELECT * FROM users WHERE username=?")
		if err != nil {
			panic(err)
		}
		rows, err := stmt.Query(payload.Username)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		if rows.Next() {
			return c.Status(fiber.StatusConflict).Send([]byte("Username already exists"))
		}

		// Insert new user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		stmt, err = db.Prepare("INSERT INTO users(username, password) values(?,?)")
		if err != nil {
			panic(err)
		}
		_, err = stmt.Exec(payload.Username, hashedPassword)
		if err != nil {
			panic(err)
		}

		// Update session
		sess, err := store.Get(c)
		if err != nil {
			panic(err)
		}
		sess.Set("username", payload.Username)
		if err := sess.Save(); err != nil {
			panic(err)
		}

		return c.SendString(fmt.Sprintf("Welcome, %v", payload.Username))
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		payload := User{}

		// Parse body et cetera
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if payload.Username == "" || payload.Password == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// Check for user
		stmt, err := db.Prepare("SELECT password FROM users WHERE username=?")
		if err != nil {
			panic(err)
		}
		row := stmt.QueryRow(payload.Username)
		var hashedPassword string
		switch err := row.Scan(&hashedPassword); err {
		case sql.ErrNoRows:
			return c.Status(fiber.StatusNotFound).Send([]byte("User not found"))
		case nil:
		default:
			panic(err)
		}

		// Compare passwords
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(payload.Password))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).Send([]byte("Invalid password"))
		}

		sess, err := store.Get(c)
		if err != nil {
			panic(err)
		}

		sess.Set("username", payload.Username)
		if err := sess.Save(); err != nil {
			panic(err)
		}

		return c.SendString(fmt.Sprintf("Welcome back, %v", payload.Username))
	})

	app.Get("/me", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			panic(err)
		}

		username := sess.Get("username")

		if username != nil {
			return c.SendString(username.(string))
		}
		return c.SendStatus(fiber.StatusUnauthorized)
	})

	app.Delete("/me", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			panic(err)
		}

		username := sess.Get("username")

		if username == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		if err := sess.Destroy(); err != nil {
			panic(err)
		}

		return c.Status(fiber.StatusOK).Send([]byte(fmt.Sprintf("Goodbye, %s", username)))
	})

	app.Listen(":3000")
}
