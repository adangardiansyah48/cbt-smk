package handlers

import (
	"context"
	"time"

	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) ExamResults(c *fiber.Ctx) error {
	examID := c.Params("id")
	rows, err := a.DB.Query(context.Background(), `
		select r.id, r.exam_id, r.student_id, r.total_pg_score, r.total_essay_score, r.final_score,
		       r.status, r.started_at, r.submitted_at, r.last_seen_at, r.cheat_warnings,
		       u.full_name, s.nisn, coalesce(cl.class_name,''),
		       (select count(*) from exam_responses er where er.exam_id=r.exam_id and er.student_id=r.student_id and er.answer is not null and er.answer <> 'null'::jsonb)
		from exam_results r
		join students s on s.id=r.student_id
		join users u on u.id=s.user_id
		left join classes cl on cl.id=s.class_id
		where r.exam_id=$1
		order by u.full_name
	`, examID)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.ExamResult{}
	for rows.Next() {
		var r models.ExamResult
		if err := rows.Scan(&r.ID, &r.ExamID, &r.StudentID, &r.TotalPGScore, &r.TotalEssayScore, &r.FinalScore,
			&r.Status, &r.StartedAt, &r.SubmittedAt, &r.LastSeenAt, &r.CheatWarnings,
			&r.StudentName, &r.NISN, &r.ClassName, &r.AnsweredCount); err != nil {
			return Fail(c, err)
		}
		out = append(out, r)
	}
	return c.JSON(out)
}

func (a *App) MonitorExam(c *fiber.Ctx) error {
	examID := c.Params("id")
	rows, err := a.DB.Query(context.Background(), `
		select r.student_id::text, u.full_name, s.nisn, coalesce(cl.class_name,''),
		       r.status, r.last_seen_at, r.cheat_warnings, r.final_score,
		       (select count(*) from exam_responses er where er.exam_id=r.exam_id and er.student_id=r.student_id and er.answer is not null and er.answer <> 'null'::jsonb) as answered,
		       (select count(*) from exam_questions eq where eq.exam_id=r.exam_id) as total
		from exam_results r
		join students s on s.id=r.student_id
		join users u on u.id=s.user_id
		left join classes cl on cl.id=s.class_id
		where r.exam_id=$1
		order by u.full_name
	`, examID)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	type row struct {
		StudentID     string    `json:"student_id"`
		StudentName   string    `json:"student_name"`
		NISN          string    `json:"nisn"`
		ClassName     string    `json:"class_name"`
		Status        string    `json:"status"`
		LastSeenAt    time.Time `json:"last_seen_at"`
		CheatWarnings int       `json:"cheat_warnings"`
		FinalScore    float64   `json:"final_score"`
		Answered      int       `json:"answered"`
		Total         int       `json:"total"`
		Online        bool      `json:"online"`
	}
	out := []row{}
	now := time.Now()
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.StudentID, &r.StudentName, &r.NISN, &r.ClassName, &r.Status, &r.LastSeenAt, &r.CheatWarnings, &r.FinalScore, &r.Answered, &r.Total); err != nil {
			return Fail(c, err)
		}
		r.Online = now.Sub(r.LastSeenAt) < 20*time.Second && r.Status == "in_progress"
		out = append(out, r)
	}
	return c.JSON(out)
}

func (a *App) GradeExam(c *fiber.Ctx) error {
	examID := c.Params("id")
	ctx := context.Background()
	rows, err := a.DB.Query(ctx, `select student_id::text from exam_results where exam_id=$1`, examID)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return Fail(c, err)
		}
		pg, essay, final, err := a.gradeStudent(ctx, examID, sid)
		if err != nil {
			continue
		}
		_, _ = a.DB.Exec(ctx, `
			update exam_results set total_pg_score=$3, total_essay_score=$4, final_score=$5,
			       status = case when status='in_progress' then status else 'graded' end
			where exam_id=$1 and student_id=$2
		`, examID, sid, pg, essay, final)
		n++
	}
	return c.JSON(fiber.Map{"graded": n})
}
