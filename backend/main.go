package main

import (
	"log"
	"os"
	"strings"

	"cbt-api/config"
	"cbt-api/handlers"
	"cbt-api/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL kosong")
	}
	pool := config.ConnectDB(cfg.DatabaseURL)
	defer pool.Close()

	app := fiber.New(fiber.Config{AppName: "CBT SMK API"})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendOrigin,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	h := &handlers.App{DB: pool}

	app.Get("/ping", h.Ping)
	app.Get("/api/school", h.SchoolProfile)
	app.Post("/api/auth/login", h.StudentLogin)
	app.Post("/api/auth/staff", h.StaffLogin)

	api := app.Group("/api", middleware.Protected())
	api.Get("/me", h.Me)

	api.Get("/exam/paper", middleware.Role("student"), h.GetExamPaper)
	api.Post("/exam/response", middleware.Role("student"), h.UpsertResponse)
	api.Post("/exam/heartbeat", middleware.Role("student"), h.Heartbeat)
	api.Post("/exam/submit", middleware.Role("student"), h.SubmitExam)

	staff := api.Group("/", middleware.Role("admin", "teacher"))
	staff.Get("/majors", h.ListMajors)
	staff.Post("/majors", h.CreateMajor)
	staff.Get("/classes", h.ListClasses)
	staff.Post("/classes", h.CreateClass)
	staff.Get("/subjects", h.ListSubjects)
	staff.Post("/subjects", h.CreateSubject)
	staff.Get("/students", h.ListStudents)
	staff.Post("/students", h.CreateStudent)
	staff.Get("/teachers", h.ListTeachers)
	staff.Post("/teachers", h.CreateTeacher)
	staff.Get("/questions", h.ListQuestions)
	staff.Post("/questions", h.CreateQuestion)
	staff.Post("/questions/import", h.ImportQuestions)
	staff.Delete("/questions/:id", h.DeleteQuestion)
	staff.Get("/exams", h.ListExams)
	staff.Post("/exams", h.CreateExam)
	staff.Post("/exams/:id/toggle", h.ToggleExam)
	staff.Get("/exams/:id/results", h.ExamResults)
	staff.Get("/exams/:id/monitor", h.MonitorExam)
	staff.Post("/exams/:id/grade", h.GradeExam)

	port := cfg.Port
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	log.Printf("listening %s", port)
	log.Fatal(app.Listen(port))
}
