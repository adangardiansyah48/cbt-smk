package handlers

import (
	"context"

	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
)

func (a *App) ListMajors(c *fiber.Ctx) error {
	rows, err := a.DB.Query(context.Background(), `select id, code, name, department_head from majors order by code`)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Major{}
	for rows.Next() {
		var m models.Major
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.DepartmentHead); err != nil {
			return Fail(c, err)
		}
		out = append(out, m)
	}
	return c.JSON(out)
}

func (a *App) CreateMajor(c *fiber.Ctx) error {
	var m models.Major
	if err := c.BodyParser(&m); err != nil {
		return Bad(c, "payload tidak valid")
	}
	err := a.DB.QueryRow(context.Background(), `
		insert into majors (code, name, department_head) values ($1,$2,$3)
		returning id
	`, m.Code, m.Name, m.DepartmentHead).Scan(&m.ID)
	if err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(m)
}

func (a *App) ListClasses(c *fiber.Ctx) error {
	rows, err := a.DB.Query(context.Background(), `
		select c.id, c.major_id, c.grade_level, c.class_name, m.code
		from classes c join majors m on m.id=c.major_id
		order by c.grade_level, c.class_name
	`)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Class{}
	for rows.Next() {
		var cl models.Class
		if err := rows.Scan(&cl.ID, &cl.MajorID, &cl.GradeLevel, &cl.ClassName, &cl.MajorCode); err != nil {
			return Fail(c, err)
		}
		out = append(out, cl)
	}
	return c.JSON(out)
}

func (a *App) CreateClass(c *fiber.Ctx) error {
	var cl models.Class
	if err := c.BodyParser(&cl); err != nil {
		return Bad(c, "payload tidak valid")
	}
	err := a.DB.QueryRow(context.Background(), `
		insert into classes (major_id, grade_level, class_name) values ($1,$2,$3)
		returning id
	`, cl.MajorID, cl.GradeLevel, cl.ClassName).Scan(&cl.ID)
	if err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(cl)
}

func (a *App) ListSubjects(c *fiber.Ctx) error {
	rows, err := a.DB.Query(context.Background(), `
		select id, major_id, code, name, pass_grade, subject_type from subjects order by name
	`)
	if err != nil {
		return Fail(c, err)
	}
	defer rows.Close()
	out := []models.Subject{}
	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.MajorID, &s.Code, &s.Name, &s.PassGrade, &s.SubjectType); err != nil {
			return Fail(c, err)
		}
		out = append(out, s)
	}
	return c.JSON(out)
}

func (a *App) CreateSubject(c *fiber.Ctx) error {
	var s models.Subject
	if err := c.BodyParser(&s); err != nil {
		return Bad(c, "payload tidak valid")
	}
	err := a.DB.QueryRow(context.Background(), `
		insert into subjects (major_id, code, name, pass_grade, subject_type)
		values ($1,$2,$3,$4,$5) returning id
	`, s.MajorID, s.Code, s.Name, s.PassGrade, s.SubjectType).Scan(&s.ID)
	if err != nil {
		return Fail(c, err)
	}
	return c.Status(201).JSON(s)
}

func (a *App) SchoolProfile(c *fiber.Ctx) error {
	var name, npsn, address, head, logo *string
	err := a.DB.QueryRow(context.Background(), `
		select school_name, npsn, address, headmaster_name, logo_url
		from school_profiles order by created_at limit 1
	`).Scan(&name, &npsn, &address, &head, &logo)
	if err != nil {
		return c.JSON(fiber.Map{"school_name": "SMK CBT"})
	}
	return c.JSON(fiber.Map{
		"school_name":     name,
		"npsn":            npsn,
		"address":         address,
		"headmaster_name": head,
		"logo_url":        logo,
	})
}
