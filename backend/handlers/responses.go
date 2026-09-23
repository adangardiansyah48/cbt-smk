package handlers

import (
	"context"
	"encoding/json"
	"time"

	"cbt-api/middleware"
	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) UpsertResponse(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.Role != "student" {
		return c.Status(403).JSON(fiber.Map{"error": "hanya siswa"})
	}
	var req models.UpsertResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.QuestionID == "" {
		return Bad(c, "question_id wajib")
	}
	if len(req.Answer) == 0 {
		req.Answer = json.RawMessage(`null`)
	}
	ctx := context.Background()

	var status string
	_ = a.DB.QueryRow(ctx, `select status from exam_results where exam_id=$1 and student_id=$2`, claims.ExamID, claims.StudentID).Scan(&status)
	if status == "submitted" || status == "graded" {
		return c.Status(403).JSON(fiber.Map{"error": "ujian sudah selesai"})
	}

	_, err := a.DB.Exec(ctx, `
		insert into exam_responses (exam_id, student_id, question_id, answer, is_flagged)
		values ($1,$2,$3,$4,$5)
		on conflict (exam_id, student_id, question_id)
		do update set answer=excluded.answer, is_flagged=excluded.is_flagged, updated_at=now()
	`, claims.ExamID, claims.StudentID, req.QuestionID, req.Answer, req.IsFlagged)
	if err != nil {
		return Fail(c, err)
	}
	_, _ = a.DB.Exec(ctx, `update exam_results set last_seen_at=now() where exam_id=$1 and student_id=$2`, claims.ExamID, claims.StudentID)
	return c.JSON(fiber.Map{"ok": true, "saved_at": time.Now()})
}

func (a *App) Heartbeat(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauth"})
	}
	warn := c.QueryInt("cheat", 0)
	if warn > 0 {
		_, _ = a.DB.Exec(context.Background(), `
			update exam_results set last_seen_at=now(), cheat_warnings = cheat_warnings + $3
			where exam_id=$1 and student_id=$2
		`, claims.ExamID, claims.StudentID, warn)
	} else {
		_, _ = a.DB.Exec(context.Background(), `
			update exam_results set last_seen_at=now() where exam_id=$1 and student_id=$2
		`, claims.ExamID, claims.StudentID)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (a *App) SubmitExam(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.Role != "student" {
		return c.Status(403).JSON(fiber.Map{"error": "hanya siswa"})
	}
	ctx := context.Background()

	pg, essay, final, err := a.gradeStudent(ctx, claims.ExamID, claims.StudentID)
	if err != nil {
		return Fail(c, err)
	}
	_, err = a.DB.Exec(ctx, `
		update exam_results
		set total_pg_score=$3, total_essay_score=$4, final_score=$5,
		    status='submitted', submitted_at=now(), last_seen_at=now()
		where exam_id=$1 and student_id=$2
	`, claims.ExamID, claims.StudentID, pg, essay, final)
	if err != nil {
		return Fail(c, err)
	}
	return c.JSON(fiber.Map{
		"ok":                true,
		"total_pg_score":    pg,
		"total_essay_score": essay,
		"final_score":       final,
	})
}

func (a *App) gradeStudent(ctx context.Context, examID, studentID string) (float64, float64, float64, error) {
	type qrow struct {
		ID   string
		Type string
		Opts json.RawMessage
		Ans  json.RawMessage
	}
	rows, err := a.DB.Query(ctx, `
		select q.id::text, q.question_type, q.options, coalesce(r.answer, 'null'::jsonb)
		from exam_questions eq
		join questions q on q.id=eq.question_id
		left join exam_responses r on r.question_id=q.id and r.exam_id=eq.exam_id and r.student_id=$2
		where eq.exam_id=$1
	`, examID, studentID)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()
	var autoItems []qrow
	essayCount := 0
	for rows.Next() {
		var r qrow
		if err := rows.Scan(&r.ID, &r.Type, &r.Opts, &r.Ans); err != nil {
			return 0, 0, 0, err
		}
		if r.Type == "essay" || r.Type == "short" {
			essayCount++
			continue
		}
		autoItems = append(autoItems, r)
	}
	correct := 0
	for _, it := range autoItems {
		if gradeAuto(it.Type, it.Opts, it.Ans) {
			correct++
		}
	}
	autoTotal := len(autoItems)
	pg := 0.0
	if autoTotal > 0 {
		pg = (float64(correct) / float64(autoTotal)) * 100
	}
	essay := 0.0
	final := pg
	if autoTotal+essayCount > 0 {
		final = (pg*float64(autoTotal) + essay*float64(essayCount)) / float64(autoTotal+essayCount)
	}
	return pg, essay, final, nil
}

func gradeAuto(qtype string, opts json.RawMessage, ans json.RawMessage) bool {
	var options []map[string]any
	_ = json.Unmarshal(opts, &options)
	var given any
	_ = json.Unmarshal(ans, &given)
	if given == nil {
		return false
	}
	switch qtype {
	case "pg":
		want := ""
		for _, o := range options {
			if b, _ := o["correct"].(bool); b {
				if id, ok := o["id"].(string); ok {
					want = id
				} else if t, ok := o["text"].(string); ok {
					want = t
				}
			}
		}
		got, _ := given.(string)
		return got != "" && got == want
	case "pg_complex":
		want := map[string]bool{}
		for _, o := range options {
			if b, _ := o["correct"].(bool); b {
				if id, ok := o["id"].(string); ok {
					want[id] = true
				}
			}
		}
		gotArr, ok := given.([]any)
		if !ok || len(gotArr) != len(want) {
			return false
		}
		for _, g := range gotArr {
			s, _ := g.(string)
			if !want[s] {
				return false
			}
		}
		return true
	case "matching":
		want := map[string]string{}
		for _, o := range options {
			left, _ := o["left"].(string)
			right, _ := o["right"].(string)
			if left != "" {
				want[left] = right
			}
		}
		gotMap, ok := given.(map[string]any)
		if !ok || len(gotMap) != len(want) {
			return false
		}
		for k, v := range want {
			gs, _ := gotMap[k].(string)
			if gs != v {
				return false
			}
		}
		return true
	default:
		return false
	}
}
