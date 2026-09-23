package handlers

import (
	"context"

	"cbt-api/middleware"
	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) StudentLogin(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.NISN == "" || req.Token == "" {
		return Bad(c, "NISN dan token ujian wajib")
	}

	ctx := context.Background()
	var examID, examTitle string
	var isPKL bool
	var isActive bool
	err := a.DB.QueryRow(ctx, `
		select id::text, title, is_pkl_allowed, is_active
		from exams where upper(token)=upper($1) limit 1
	`, req.Token).Scan(&examID, &examTitle, &isPKL, &isActive)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "token ujian tidak ditemukan"})
	}
	if !isActive {
		return c.Status(403).JSON(fiber.Map{"error": "ujian belum aktif"})
	}

	var student models.Student
	var className, majorCode *string
	err = a.DB.QueryRow(ctx, `
		select s.id, s.user_id, s.nisn, s.nis, s.class_id, s.is_pkl_active, u.full_name,
		       c.class_name, m.code
		from students s
		join users u on u.id = s.user_id
		left join classes c on c.id = s.class_id
		left join majors m on m.id = c.major_id
		where s.nisn = $1
	`, req.NISN).Scan(
		&student.ID, &student.UserID, &student.NISN, &student.NIS, &student.ClassID,
		&student.IsPKLActive, &student.FullName, &className, &majorCode,
	)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "NISN tidak terdaftar"})
	}
	if className != nil {
		student.ClassName = *className
	}
	if majorCode != nil {
		student.MajorCode = *majorCode
	}

	if student.IsPKLActive && !isPKL {
		return c.Status(403).JSON(fiber.Map{"error": "siswa PKL tidak diizinkan pada ujian ini"})
	}

	if student.ClassID != nil {
		var ok int
		err = a.DB.QueryRow(ctx, `
			select count(*) from exam_classes where exam_id=$1 and class_id=$2
		`, examID, *student.ClassID).Scan(&ok)
		if err != nil {
			return Fail(c, err)
		}
		if ok == 0 {
			return c.Status(403).JSON(fiber.Map{"error": "kelas kamu tidak termasuk target ujian"})
		}
	}

	_, err = a.DB.Exec(ctx, `
		insert into exam_results (exam_id, student_id, status)
		values ($1, $2, 'in_progress')
		on conflict (exam_id, student_id) do update set last_seen_at = now()
	`, examID, student.ID)
	if err != nil {
		return Fail(c, err)
	}

	token, err := middleware.Sign(models.Claims{
		UserID:    student.UserID.String(),
		StudentID: student.ID.String(),
		ExamID:    examID,
		Role:      "student",
		FullName:  student.FullName,
	})
	if err != nil {
		return Fail(c, err)
	}

	return c.JSON(fiber.Map{
		"token": token,
		"role":  "student",
		"exam":  fiber.Map{"id": examID, "title": examTitle},
		"student": student,
	})
}

func (a *App) StaffLogin(c *fiber.Ctx) error {
	var req models.StaffLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return Bad(c, "payload tidak valid")
	}
	if req.Username == "" || req.Password == "" {
		return Bad(c, "username dan password wajib")
	}

	ctx := context.Background()
	var user models.User
	var storedHash string
	err := a.DB.QueryRow(ctx, `
		select id, username, full_name, role, coalesce(password_hash, '')
		from users where username=$1
	`, req.Username).Scan(&user.ID, &user.Username, &user.FullName, &user.Role, &storedHash)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "akun tidak ditemukan"})
	}
	if user.Role == "student" {
		return c.Status(403).JSON(fiber.Map{"error": "siswa login via NISN + token"})
	}
	if storedHash == "" || storedHash != HashPassword(req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "password salah"})
	}

	claims := models.Claims{
		UserID:   user.ID.String(),
		Role:     user.Role,
		FullName: user.FullName,
	}
	if user.Role == "teacher" {
		var tid string
		_ = a.DB.QueryRow(ctx, `select id::text from teachers where user_id=$1`, user.ID).Scan(&tid)
		claims.TeacherID = tid
	}

	token, err := middleware.Sign(claims)
	if err != nil {
		return Fail(c, err)
	}
	return c.JSON(fiber.Map{"token": token, "role": user.Role, "user": user})
}

func (a *App) Me(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	return c.JSON(claims)
}

func (a *App) Ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true, "service": "cbt-api"})
}
