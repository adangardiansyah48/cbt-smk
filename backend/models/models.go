package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID  `json:"id"`
	AuthID   *uuid.UUID `json:"auth_id,omitempty"`
	Username string     `json:"username"`
	FullName string     `json:"full_name"`
	Role     string     `json:"role"`
}

type Student struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	NISN        string    `json:"nisn"`
	NIS         *string   `json:"nis,omitempty"`
	ClassID     *uuid.UUID `json:"class_id,omitempty"`
	IsPKLActive bool      `json:"is_pkl_active"`
	FullName    string    `json:"full_name,omitempty"`
	ClassName   string    `json:"class_name,omitempty"`
	MajorCode   string    `json:"major_code,omitempty"`
}

type Teacher struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	NIP        *string   `json:"nip_nuptk,omitempty"`
	IsInternal bool      `json:"is_internal"`
	Gender     *string   `json:"gender,omitempty"`
	FullName   string    `json:"full_name,omitempty"`
	Username   string    `json:"username,omitempty"`
}

type Major struct {
	ID             uuid.UUID `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	DepartmentHead *string   `json:"department_head,omitempty"`
}

type Class struct {
	ID         uuid.UUID `json:"id"`
	MajorID    uuid.UUID `json:"major_id"`
	GradeLevel int       `json:"grade_level"`
	ClassName  string    `json:"class_name"`
	MajorCode  string    `json:"major_code,omitempty"`
}

type Subject struct {
	ID          uuid.UUID  `json:"id"`
	MajorID     *uuid.UUID `json:"major_id,omitempty"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	PassGrade   float64    `json:"pass_grade"`
	SubjectType string     `json:"subject_type"`
}

type Question struct {
	ID           uuid.UUID       `json:"id"`
	SubjectID    uuid.UUID       `json:"subject_id"`
	TeacherID    *uuid.UUID      `json:"teacher_id,omitempty"`
	QuestionType string          `json:"question_type"`
	Content      string          `json:"content"`
	MediaURL     *string         `json:"media_url,omitempty"`
	MediaType    *string         `json:"media_type,omitempty"`
	Options      json.RawMessage `json:"options"`
}

type Exam struct {
	ID                uuid.UUID  `json:"id"`
	Title             string     `json:"title"`
	SubjectID         uuid.UUID  `json:"subject_id"`
	SemesterID        *uuid.UUID `json:"semester_id,omitempty"`
	Duration          int        `json:"duration"`
	Token             string     `json:"token"`
	IsPKLAllowed      bool       `json:"is_pkl_allowed"`
	ShuffleQuestions  bool       `json:"shuffle_questions"`
	ShuffleOptions    bool       `json:"shuffle_options"`
	StartAt           *time.Time `json:"start_at,omitempty"`
	EndAt             *time.Time `json:"end_at,omitempty"`
	IsActive          bool       `json:"is_active"`
	SubjectName       string     `json:"subject_name,omitempty"`
	QuestionCount     int        `json:"question_count,omitempty"`
}

type ExamQuestion struct {
	Question
	OrderNo int `json:"order_no"`
}

type ExamResponse struct {
	ID         uuid.UUID       `json:"id"`
	ExamID     uuid.UUID       `json:"exam_id"`
	StudentID  uuid.UUID       `json:"student_id"`
	QuestionID uuid.UUID       `json:"question_id"`
	Answer     json.RawMessage `json:"answer"`
	IsFlagged  bool            `json:"is_flagged"`
}

type ExamResult struct {
	ID              uuid.UUID  `json:"id"`
	ExamID          uuid.UUID  `json:"exam_id"`
	StudentID       uuid.UUID  `json:"student_id"`
	TotalPGScore    float64    `json:"total_pg_score"`
	TotalEssayScore float64    `json:"total_essay_score"`
	FinalScore      float64    `json:"final_score"`
	Status          string     `json:"status"`
	StartedAt       time.Time  `json:"started_at"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	LastSeenAt      time.Time  `json:"last_seen_at"`
	CheatWarnings   int        `json:"cheat_warnings"`
	StudentName     string     `json:"student_name,omitempty"`
	NISN            string     `json:"nisn,omitempty"`
	ClassName       string     `json:"class_name,omitempty"`
	AnsweredCount   int        `json:"answered_count,omitempty"`
}

type Claims struct {
	UserID    string `json:"user_id"`
	StudentID string `json:"student_id,omitempty"`
	TeacherID string `json:"teacher_id,omitempty"`
	ExamID    string `json:"exam_id,omitempty"`
	Role      string `json:"role"`
	FullName  string `json:"full_name"`
}

type LoginRequest struct {
	NISN  string `json:"nisn"`
	Token string `json:"token"`
}

type StaffLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpsertResponseRequest struct {
	QuestionID string          `json:"question_id"`
	Answer     json.RawMessage `json:"answer"`
	IsFlagged  bool            `json:"is_flagged"`
}

type CreateExamRequest struct {
	Title            string   `json:"title"`
	SubjectID        string   `json:"subject_id"`
	SemesterID       *string  `json:"semester_id"`
	Duration         int      `json:"duration"`
	IsPKLAllowed     bool     `json:"is_pkl_allowed"`
	ShuffleQuestions bool     `json:"shuffle_questions"`
	ShuffleOptions   bool     `json:"shuffle_options"`
	StartAt          *string  `json:"start_at"`
	EndAt            *string  `json:"end_at"`
	IsActive         bool     `json:"is_active"`
	ClassIDs         []string `json:"class_ids"`
	QuestionIDs      []string `json:"question_ids"`
}

type CreateQuestionRequest struct {
	SubjectID    string          `json:"subject_id"`
	QuestionType string          `json:"question_type"`
	Content      string          `json:"content"`
	MediaURL     *string         `json:"media_url"`
	MediaType    *string         `json:"media_type"`
	Options      json.RawMessage `json:"options"`
}

type CreateStudentRequest struct {
	Username    string `json:"username"`
	FullName    string `json:"full_name"`
	NISN        string `json:"nisn"`
	NIS         string `json:"nis"`
	ClassID     string `json:"class_id"`
	IsPKLActive bool   `json:"is_pkl_active"`
}

type CreateTeacherRequest struct {
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	Password   string `json:"password"`
	NIP        string `json:"nip_nuptk"`
	IsInternal bool   `json:"is_internal"`
	Gender     string `json:"gender"`
}

type ImportQuestionRow struct {
	QuestionType string          `json:"question_type"`
	Content      string          `json:"content"`
	MediaURL     *string         `json:"media_url"`
	MediaType    *string         `json:"media_type"`
	Options      json.RawMessage `json:"options"`
}
