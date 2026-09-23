package handlers

import (
	"context"

	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) ListStudents(c *fiber.Ctx) error {
	rows, err := a.DB.Query(context.Background(), `
		select s.id, s.user_id, s.nisn, s.nis, s.class_id, s.is_pkl_active,
		       u.full_name, coalesce(cl.class_name,''), coalesce(m.code,'')
		from students s
		join users u on u.id=s.user_id
		left join classes cl on cl.id=s.class_id
		left join majors m on m.id=cl.major_id
		order by u.full_name
	`)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Student{}
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NISN, &s.NIS, &s.ClassID, &s.IsPKLActive, &s.FullName, &s.ClassName, &s.MajorCode); err != nil {
			return Fail(c, err)
		}
		out = append(out, s)
	}
	return c.JSON(out)
}

func (a *App) CreateStudent(c *fiber.Ctx) error {
	var req models.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.NISN == "" || req.FullName == "" {
		return Bad(c, "nama dan NISN wajib")
	}
	username := req.Username
	if username == "" {
		username = req.NISN
	}
	ctx := context.Background()
	tx, err := a.DB.Begin(ctx)
	if err != nil {
		return Fail(c, err)
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		insert into users (username, full_name, role, password_hash)
		values ($1,$2,'student',$3) returning id::text
	`, username, req.FullName, HashPassword(req.NISN)).Scan(&userID)
	if err != nil {
		return Fail(c, err)
	}

	var classID interface{}
	if req.ClassID != "" {
		classID = req.ClassID
	}
	var nis interface{}
	if req.NIS != "" {
		nis = req.NIS
	}
	var sid string
	err = tx.QueryRow(ctx, `
		insert into students (user_id, nisn, nis, class_id, is_pkl_active)
		values ($1,$2,$3,$4,$5) returning id::text
	`, userID, req.NISN, nis, classID, req.IsPKLActive).Scan(&sid)
	if err != nil {
		return Fail(c, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"id": sid, "user_id": userID, "nisn": req.NISN})
}

func (a *App) ListTeachers(c *fiber.Ctx) error {
	rows, err := a.DB.Query(context.Background(), `
		select t.id, t.user_id, t.nip_nuptk, t.is_internal, t.gender, u.full_name, u.username
		from teachers t join users u on u.id=t.user_id
		order by u.full_name
	`)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Teacher{}
	for rows.Next() {
		var t models.Teacher
		if err := rows.Scan(&t.ID, &t.UserID, &t.NIP, &t.IsInternal, &t.Gender, &t.FullName, &t.Username); err != nil {
			return Fail(c, err)
		}
		out = append(out, t)
	}
	return c.JSON(out)
}

func (a *App) CreateTeacher(c *fiber.Ctx) error {
	var req models.CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.Username == "" || req.FullName == "" || req.Password == "" {
		return Bad(c, "username, nama, password wajib")
	}
	ctx := context.Background()
	tx, err := a.DB.Begin(ctx)
	if err != nil {
		return Fail(c, err)
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		insert into users (username, full_name, role, password_hash)
		values ($1,$2,'teacher',$3) returning id::text
	`, req.Username, req.FullName, HashPassword(req.Password)).Scan(&userID)
	if err != nil {
		return Fail(c, err)
	}
	var nip, gender interface{}
	if req.NIP != "" {
		nip = req.NIP
	}
	if req.Gender != "" {
		gender = req.Gender
	}
	var tid string
	err = tx.QueryRow(ctx, `
		insert into teachers (user_id, nip_nuptk, is_internal, gender)
		values ($1,$2,$3,$4) returning id::text
	`, userID, nip, req.IsInternal, gender).Scan(&tid)
	if err != nil {
		return Fail(c, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"id": tid, "user_id": userID})
}
