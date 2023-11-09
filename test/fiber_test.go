package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "embed"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache/v2"
	"github.com/stretchr/testify/assert"
)

var engine = mustache.New("../template", ".mustache")

// tambahkan fiber.Config jika ingin buat error handler
var app = fiber.New(fiber.Config{
	Views: engine,
	ErrorHandler: func(ctx *fiber.Ctx, err error) error {
		ctx.Status(fiber.StatusInternalServerError)
		return ctx.SendString("Error: " + err.Error())
	},
})

func TestRoutingHelloWorld(t *testing.T) {
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello world")
	})

	request := httptest.NewRequest("GET", "/", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello world", string(bytes))
}

func TestCtx(t *testing.T) {
	app.Get("/hello", func(ctx *fiber.Ctx) error {
		// default jika query parameter name kosong = Guest
		name := ctx.Query("name", "Guest")
		return ctx.SendString("Hello " + name)
	})

	request := httptest.NewRequest("GET", "/hello?name=Dicki", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Dicki", string(bytes))
}

func TestHttpRequest(t *testing.T) {
	app.Get("/request", func(ctx *fiber.Ctx) error {
		first := ctx.Get("firstname")
		last := ctx.Cookies("lastname")
		return ctx.SendString("Hello " + first + " " + last)
	})

	request := httptest.NewRequest("GET", "/request", nil)
	request.Header.Set("firstname", "Dicki")
	request.AddCookie(&http.Cookie{Name: "lastname", Value: "Darmawan Saputra"})
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Dicki Darmawan Saputra", string(bytes))
}

func TestRouteParameter(t *testing.T) {
	app.Get("/users/:userId/orders/:orderId", func(ctx *fiber.Ctx) error {
		userId := ctx.Params("userId")
		orderId := ctx.Params("orderId")
		return ctx.SendString("Get order " + orderId + " from " + userId)
	})

	request := httptest.NewRequest("GET", "/users/dicki/orders/10", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Get order 10 from dicki", string(bytes))
}

func TestFormRequest(t *testing.T) {
	app.Post("/form", func(ctx *fiber.Ctx) error {
		name := ctx.FormValue("name")
		return ctx.SendString("Hello " + name)
	})

	body := strings.NewReader("name=Dicki")
	request := httptest.NewRequest("POST", "/form", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Dicki", string(bytes))
}

//go:embed source/contoh.txt
var contohFile []byte

func TestFormUpload(t *testing.T) {
	app.Post("/upload", func(ctx *fiber.Ctx) error {
		file, err := ctx.FormFile("file")
		if err != nil {
			return err
		}

		err = ctx.SaveFile(file, "../target/"+file.Filename)
		if err != nil {
			return err
		}

		return ctx.SendString("Upload success")
	})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("file", "contoh.txt")
	assert.Nil(t, err)
	file.Write(contohFile)
	writer.Close()

	request := httptest.NewRequest("POST", "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Upload success", string(bytes))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T) {
	app.Post("/login", func(ctx *fiber.Ctx) error {
		body := ctx.Body()
		// gunakan new agar jadi pointer
		request := new(LoginRequest)
		err := json.Unmarshal(body, request)
		if err != nil {
			return err
		}

		return ctx.SendString("Hello " + request.Username)
	})

	body := strings.NewReader(`{"username":"dicki","password":"rahasia"}`)
	request := httptest.NewRequest("POST", "/login", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello dicki", string(bytes))
}

type RegisterRequest struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
	Name     string `json:"name" xml:"name" form:"name"`
}

func TestBodyParser(t *testing.T) {
	app.Post("/register", func(ctx *fiber.Ctx) error {
		// gunakan new agar jadi pointer
		request := new(RegisterRequest)
		err := ctx.BodyParser(request)
		if err != nil {
			return err
		}

		return ctx.SendString("Register success " + request.Name)
	})
}

func TestBodyParserJson(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`{"username":"dicki","password":"rahasia","name":"Dicki Darmawan Saputra"}`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register success Dicki Darmawan Saputra", string(bytes))
}

func TestBodyParserForm(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader("username=dicki&password=rahasia&name=Dicki+Darmawan+Saputra")
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register success Dicki Darmawan Saputra", string(bytes))
}

func TestBodyParserXml(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(
		`<RegisterRequest>
			<username>dicki</username>
			<password>rahasia</password>
			<name>Dicki Darmawan Saputra</name>
		</RegisterRequest>`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/xml")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register success Dicki Darmawan Saputra", string(bytes))
}

func TestResponseJson(t *testing.T) {
	app.Get("/user", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"username": "dicki",
			"name":     "Dicki Darmawan Saputra",
		})
	})

	request := httptest.NewRequest("GET", "/user", nil)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	// hati-hati harus sesuai abjad urutan jsonnya
	assert.Equal(t, `{"name":"Dicki Darmawan Saputra","username":"dicki"}`, string(bytes))
}

func TestDownloadFile(t *testing.T) {
	app.Get("/download", func(ctx *fiber.Ctx) error {
		return ctx.Download("../source/contoh.txt", "contoh.txt")
	})

	request := httptest.NewRequest("GET", "/download", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, `attachment; filename="contoh.txt"`, response.Header.Get("Content-Disposition"))

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "contoh file yang diupload", string(bytes))
}

func TestRoutingGroup(t *testing.T) {
	helloWorld := func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello world")
	}

	api := app.Group("/api")
	api.Get("/hello", helloWorld)
	api.Post("/hello", helloWorld)
	api.Put("/hello", helloWorld)
	api.Delete("/hello", helloWorld)

	web := app.Group("/web")
	web.Get("/hello", helloWorld)
	web.Post("/hello", helloWorld)
	web.Put("/hello", helloWorld)
	web.Delete("/hello", helloWorld)

	request := httptest.NewRequest("GET", "/api/hello", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello world", string(bytes))
}

func TestStatic(t *testing.T) {
	app.Static("public", "../source")

	request := httptest.NewRequest("GET", "/public/contoh.txt", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "contoh file yang diupload", string(bytes))
}

func TestErrorHandling(t *testing.T) {
	app.Get("/error", func(ctx *fiber.Ctx) error {
		return errors.New("Oops")
	})

	request := httptest.NewRequest("GET", "/error", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Error: Oops", string(bytes))
}

func TestView(t *testing.T) {
	app.Get("/view", func(ctx *fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{
			"title":   "hello title",
			"header":  "hello header",
			"content": "hello content",
		})
	})

	request := httptest.NewRequest("GET", "/view", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), "hello title")
	assert.Contains(t, string(bytes), "hello header")
	assert.Contains(t, string(bytes), "hello content")
}

func TestHttpClient(t *testing.T) {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	agent := client.Get("https://example.com/")
	status, response, err := agent.String()
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, response, "Example Domain")

	fmt.Println(response)
}
