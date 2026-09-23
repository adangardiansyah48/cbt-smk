package handlers

import (
	"context"
	"encoding/json"
	"time"

	"cbt-api/middleware"
	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) ListExams(c *fiber.Ctx) error {
	rows, err := a.DB.Query(context.Background(), `
		select e.id, e.title, e.subject_id, e.semester_id, e.duration, e.token,
		       e.is_pkl_allowed, e.shuffle_questions, e.shuffle_options,
		       e.start_at, e.end_at, e.is_active, coalesce(s.name,''),
		       (select count(*) from exam_questions eq where eq.exam_id=e.id)
		from exams e
		left join subjects s on s.id=e.subject_id
		order by e.created_at desc
	`)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Exam{}
	for rows.Next() {
		var e models.Exam
		if err := rows.Scan(&e.ID, &e.Title, &e.SubjectID, &e.SemesterID, &e.Duration, &e.Token,
			&e.IsPKLAllowed, &e.ShuffleQuestions, &e.ShuffleOptions, &e.StartAt, &e.EndAt, &e.IsActive,
			&e.SubjectName, &e.QuestionCount); err != nil {
			return Fail(c, err)
		}
		out = append(out, e)
	}
	return c.JSON(out)
}

func (a *App) CreateExam(c *fiber.Ctx) error {
	var req models.CreateExamRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.Title == "" || req.SubjectID == "" {
		return Bad(c, "judul dan mapel wajib")
	}
	if req.Duration <= 0 {
		req.Duration = 90
	}
	token := RandomToken(6)
	claims := middleware.GetClaims(c)
	var createdBy any
	if claims != nil {
		createdBy = claims.UserID
	}
	var startAt, endAt, semester any
	if req.StartAt != nil && *req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, *req.StartAt)
		if err == nil {
			startAt = t
		}
	}
	if req.EndAt != nil && *req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err == nil {
			endAt = t
		}
	}
	if req.SemesterID != nil && *req.SemesterID != "" {
		semester = *req.SemesterID
	}

	ctx := context.Background()
	tx, err := a.DB.Begin(ctx)
	if err != nil {
		return Fail(c, err)
	}
	defer tx.Rollback(ctx)

	var examID string
	err = tx.QueryRow(ctx, `
		insert into exams (title, subject_id, semester_id, duration, token, is_pkl_allowed,
		                   shuffle_questions, shuffle_options, start_at, end_at, is_active, created_by)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) returning id::text
	`, req.Title, req.SubjectID, semester, req.Duration, token, req.IsPKLAllowed,
		req.ShuffleQuestions, req.ShuffleOptions, startAt, endAt, req.IsActive, createdBy).Scan(&examID)
	if err != nil {
		return Fail(c, err)
	}
	for _, cid := range req.ClassIDs {
		if _, err := tx.Exec(ctx, `insert into exam_classes (exam_id, class_id) values ($1,$2) on conflict do nothing`, examID, cid); err != nil {
			return Fail(c, err)
		}
	}
	for i, qid := range req.QuestionIDs {
		if _, err := tx.Exec(ctx, `insert into exam_questions (exam_id, question_id, order_no) values ($1,$2,$3) on conflict do nothing`, examID, qid, i+1); err != nil {
			return Fail(c, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"id": examID, "token": token})
}

func (a *App) ToggleExam(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := a.DB.Exec(context.Background(), `update exams set is_active = not is_active where id=$1`, id)
	if err != nil {
		return Fail(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func stripAnswerKey(raw json.RawMessage, shuffle bool) json.RawMessage {
	var opts []map[string]any
	if err := json.Unmarshal(raw, &opts); err != nil || len(opts) == 0 {
		return json.RawMessage(`[]`)
	}
	clean := make([]map[string]any, 0, len(opts))
	for _, o := range opts {
		item := map[string]any{}
		if v, ok := o["id"]; ok {
			item["id"] = v
		}
		if v, ok := o["text"]; ok {
			item["text"] = v
		}
		if v, ok := o["left"]; ok {
			item["left"] = v
		}
		if v, ok := o["right"]; ok {
			item["right"] = v
		}
		clean = append(clean, item)
	}
	if shuffle {
		clean = Shuffle(clean)
	}
	b, _ := json.Marshal(clean)
	return b
}

func (a *App) GetExamPaper(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.ExamID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "sesi ujian tidak ada"})
	}
	ctx := context.Background()

	var exam models.Exam
	err := a.DB.QueryRow(ctx, `
		select e.id, e.title, e.subject_id, e.semester_id, e.duration, e.token,
		       e.is_pkl_allowed, e.shuffle_questions, e.shuffle_options,
		       e.start_at, e.end_at, e.is_active, coalesce(s.name,'')
		from exams e left join subjects s on s.id=e.subject_id
		where e.id=$1
	`, claims.ExamID).Scan(&exam.ID, &exam.Title, &exam.SubjectID, &exam.SemesterID, &exam.Duration, &exam.Token,
		&exam.IsPKLAllowed, &exam.ShuffleQuestions, &exam.ShuffleOptions, &exam.StartAt, &exam.EndAt, &exam.IsActive, &exam.SubjectName)
	if err != nil {
		return Fail(c, err)
	}

	var startedAt time.Time
	var status string
	_ = a.DB.QueryRow(ctx, `
		select started_at, status from exam_results where exam_id=$1 and student_id=$2
	`, claims.ExamID, claims.StudentID).Scan(&startedAt, &status)
	if status == "submitted" || status == "graded" {
		return c.Status(403).JSON(fiber.Map{"error": "ujian sudah dikumpulkan"})
	}

	rows, err := a.DB.Query(ctx, `
		select q.id, q.subject_id, q.teacher_id, q.question_type, q.content, q.media_url, q.media_type, q.options, eq.order_no
		from exam_questions eq
		join questions q on q.id=eq.question_id
		where eq.exam_id=$1
		order by eq.order_no
	`, claims.ExamID)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	qs := []models.ExamQuestion{}
	for rows.Next() {
		var q models.ExamQuestion
		if err := rows.Scan(&q.ID, &q.SubjectID, &q.TeacherID, &q.QuestionType, &q.Content, &q.MediaURL, &q.MediaType, &q.Options, &q.OrderNo); err != nil {
			return Fail(c, err)
		}
		q.Options = stripAnswerKey(q.Options, exam.ShuffleOptions)
		qs = append(qs, q)
	}
	if exam.ShuffleQuestions {
		qs = Shuffle(qs)
	}

	saved := map[string]any{}
	r2, err := a.DB.Query(ctx, `
		select question_id::text, answer, is_flagged from exam_responses
		where exam_id=$1 and student_id=$2
	`, claims.ExamID, claims.StudentID)
	if err == nil {
		defer r2.Close()
		for r2.Next() {
			var qid string
			var ans json.RawMessage
			var flagged bool
			if err := r2.Scan(&qid, &ans, &flagged); err == nil {
				saved[qid] = fiber.Map{"answer": ans, "is_flagged": flagged}
			}
		}
	}

	remain := exam.Duration * 60
	if !startedAt.IsZero() {
		elapsed := int(time.Since(startedAt).Seconds())
		remain = exam.Duration*60 - elapsed
		if remain < 0 {
			remain = 0
		}
	}

	return c.JSON(fiber.Map{
		"exam":            exam,
		"questions":       qs,
		"saved":           saved,
		"remaining_sec":   remain,
		"started_at":      startedAt,
	})
}
