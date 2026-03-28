package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	Name      string             `bson:"name" json:"name"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Note struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title         string             `bson:"title" json:"title"`
	Content       string             `bson:"content" json:"content"`
	Summary       string             `bson:"summary" json:"summary"`
	KeyTerms      []string           `bson:"key_terms,omitempty" json:"key_terms"`
	BulletPoints  []string           `bson:"bullet_points,omitempty" json:"bullet_points"`
	ExamReadyText string             `bson:"exam_ready_text,omitempty" json:"exam_ready_text"`
	Tags          []string           `bson:"tags" json:"tags"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type Flashcard struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	NoteID     primitive.ObjectID `bson:"note_id,omitempty" json:"note_id"`
	Front      string             `bson:"front" json:"front"`
	Back       string             `bson:"back" json:"back"`
	NextReview time.Time          `bson:"next_review" json:"next_review"`
	Interval   int                `bson:"interval" json:"interval"` // days
	EaseFactor float64            `bson:"ease_factor" json:"ease_factor"`
	Repetition int                `bson:"repetition" json:"repetition"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type Quiz struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	NoteID    primitive.ObjectID `bson:"note_id,omitempty" json:"note_id"`
	Title     string             `bson:"title" json:"title"`
	Questions []Question         `bson:"questions" json:"questions"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Question struct {
	ID            string   `bson:"id" json:"id"`
	Text          string   `bson:"text" json:"text"`
	Type          string   `bson:"type" json:"type"` // mcq, true_false, fill_blank
	Options       []string `bson:"options,omitempty" json:"options"`
	CorrectAnswer string   `bson:"correct_answer" json:"correct_answer"`
	Topic         string   `bson:"topic,omitempty" json:"topic"`
}

type StudySession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Duration  int                `bson:"duration" json:"duration"` // seconds
	Type      string             `bson:"type" json:"type"`         // pomodoro, break
	Topic     string             `bson:"topic,omitempty" json:"topic"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type QuizAttempt struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	QuizID          primitive.ObjectID `bson:"quiz_id" json:"quiz_id"`
	Score           float64            `bson:"score" json:"score"`
	IncorrectTopics []string           `bson:"incorrect_topics,omitempty" json:"incorrect_topics"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}
