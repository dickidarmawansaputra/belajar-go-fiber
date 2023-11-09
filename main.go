package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New(fiber.Config{
		IdleTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		Prefork:      true,
	})

	// tambahkan prefix jika ingin middleware berjalan di prefix tertentu
	app.Use("/api", func(ctx *fiber.Ctx) error {
		fmt.Println("i'm middleware before process")
		err := ctx.Next()
		fmt.Println("i'm middleware after process")
		return err
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello wolrd")
	})

	// check prefork process
	if fiber.IsChild() {
		fmt.Println("i'm child process")
	} else {
		fmt.Println("i'm parent process")
	}

	err := app.Listen("localhost:3000")
	if err != nil {
		panic(err)
	}
}
