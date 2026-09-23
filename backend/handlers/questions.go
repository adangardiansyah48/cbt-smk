package handlers

import (
	"context"
	"encoding/json"

	"cbt-api/middleware"
	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) ListQuestions(c *fiber.Ctx) error {
	subjectID := c.Query("subject_id")
	q := `select id, subject_id, teacher_id, question_type, content, media_url, media_type, options from questions`
	args := []any{}
	if subjectID != "" {
		q += ` where subject_id=$1`
		args = append(args, subjectID)
	}
	q += ` order by created_at desc`
	rows, err := a.DB.Query(context.Background(), q, args...)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Question{}
	for rows.Next() {
		var item models.Question
		if err := rows.Scan(&item.ID, &item.SubjectID, &item.TeacherID, &item.QuestionType, &item.Content, &item.MediaURL, &item.MediaType, &item.Options); err != nil {
			return Fail(c, err)
		}
		out = append(out, item)
	}
	return c.JSON(out)
}

func (a *App) CreateQuestion(c *fiber.Ctx) error {
	var req models.CreateQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.Content == "" || req.SubjectID == "" || req.QuestionType == "" {
		return Bad(c, "subject, tipe, dan isi soal wajib")
	}
	if len(req.Options) == 0 {
		req.Options = json.RawMessage(`[]`)
	}
	claims := middleware.GetClaims(c)
	var teacherID any
	if claims != nil && claims.TeacherID != "" {
		teacherID = claims.TeacherID
	}
	var id string
	err := a.DB.QueryRow(context.Background(), `
		insert into questions (subject_id, teacher_id, question_type, content, media_url, media_type, options)
		values ($1,$2,$3,$4,$5,$6,$7) returning id::text
	`, req.SubjectID, teacherID, req.QuestionType, req.Content, req.MediaURL, req.MediaType, req.Options).Scan(&id)
	if err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"id": id})
}

func (a *App) ImportQuestions(c *fiber.Ctx) error {
	subjectID := c.Query("subject_id")
	if subjectID == "" {
		return Bad(c, "subject_id wajib")
	}
	var rows []models.ImportQuestionRow
	if err := c.BodyParser(&rows); err != nil {
		return Bad(c, "JSON array soal tidak valid")
	}
	claims := middleware.GetClaims(c)
	var teacherID any
	if claims != nil && claims.TeacherID != "" {
		teacherID = claims.TeacherID
	}
	ctx := context.Background()
	ok := 0
	for _, r := range rows {
		if r.Content == "" {
			continue
		}
		if r.QuestionType == "" {
			r.QuestionType = "pg"
		}
		if len(r.Options) == 0 {
			r.Options = json.RawMessage(`[]`)
		}
		_, err := a.DB.Exec(ctx, `
			insert into questions (subject_id, teacher_id, question_type, content, media_url, media_type, options)
			values ($1,$2,$3,$4,$5,$6,$7)
		`, subjectID, teacherID, r.QuestionType, r.Content, r.MediaURL, r.MediaType, r.Options)
		if err == nil {
			ok++
		}
	}
	return c.JSON(fiber.Map{"imported": ok})
}

func (a *App) DeleteQuestion(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := a.DB.Exec(context.Background(), `delete from questions where id=$1`, id)
	if err != nil {
		return Fail(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}
