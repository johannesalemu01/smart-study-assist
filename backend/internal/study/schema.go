package study

import (
	"context"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
)

func BuildSchema(svc *Service) (graphql.Schema, error) {
	acronymType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AcronymSuggestion",
		Fields: graphql.Fields{
			"acronym":           &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"score":             &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"readability_score": &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"familiarity_score": &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"mnemonic":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"source_words":      &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	smartNoteType := graphql.NewObject(graphql.ObjectConfig{
		Name: "SmartNotesResult",
		Fields: graphql.Fields{
			"summary":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"key_terms":       &graphql.Field{Type: graphql.NewList(graphql.String)},
			"bullet_points":   &graphql.Field{Type: graphql.NewList(graphql.String)},
			"exam_ready_text": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	questionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Question",
		Fields: graphql.Fields{
			"id":             &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"text":           &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"type":           &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"options":        &graphql.Field{Type: graphql.NewList(graphql.String)},
			"correct_answer": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"topic":          &graphql.Field{Type: graphql.String},
		},
	})

	noteType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Note",
		Fields: graphql.Fields{
			"id":              &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"title":           &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"content":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"summary":         &graphql.Field{Type: graphql.String},
			"key_terms":       &graphql.Field{Type: graphql.NewList(graphql.String)},
			"bullet_points":   &graphql.Field{Type: graphql.NewList(graphql.String)},
			"exam_ready_text": &graphql.Field{Type: graphql.String},
			"tags":            &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	flashcardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Flashcard",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"note_id":     &graphql.Field{Type: graphql.String, Resolve: oidResolver("NoteID")},
			"front":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"back":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"next_review": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("NextReview")},
			"interval":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"ease_factor": &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"repetition":  &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})

	quizType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Quiz",
		Fields: graphql.Fields{
			"id":         &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"title":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"questions":  &graphql.Field{Type: graphql.NewList(questionType)},
			"created_at": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("CreatedAt")},
		},
	})

	quizAttemptType := graphql.NewObject(graphql.ObjectConfig{
		Name: "QuizAttempt",
		Fields: graphql.Fields{
			"id":               &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"score":            &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"incorrect_topics": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"created_at":       &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("CreatedAt")},
		},
	})

	dashboardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Dashboard",
		Fields: graphql.Fields{
			"total_study_seconds": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"quiz_attempts":       &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"average_quiz_score":  &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"topics_covered":      &graphql.Field{Type: graphql.NewList(graphql.String)},
			"weak_areas":          &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: func(p graphql.ResolveParams) (any, error) { return "ok", nil }},
			"generate_acronyms": {
				Type: graphql.NewList(acronymType),
				Args: graphql.FieldConfigArgument{
					"words": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					words, _ := p.Args["words"].([]any)
					items := make([]string, 0, len(words))
					for _, w := range words {
						items = append(items, fmt.Sprintf("%v", w))
					}
					return svc.GenerateAcronyms(items), nil
				},
			},
			"smart_notes": {
				Type: smartNoteType,
				Args: graphql.FieldConfigArgument{"raw": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					raw, _ := p.Args["raw"].(string)
					return svc.SmartNotes(raw), nil
				},
			},
			"notes": {
				Type: graphql.NewList(noteType),
				Args: graphql.FieldConfigArgument{"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					return svc.ListNotes(context.Background(), uid)
				},
			},
			"flashcards": {
				Type: graphql.NewList(flashcardType),
				Args: graphql.FieldConfigArgument{
					"user_id":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"due_only": &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					dueOnly, _ := p.Args["due_only"].(bool)
					return svc.ListFlashcards(context.Background(), uid, dueOnly)
				},
			},
			"quizzes": {
				Type: graphql.NewList(quizType),
				Args: graphql.FieldConfigArgument{"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					return svc.ListQuizzes(context.Background(), uid)
				},
			},
			"dashboard": {
				Type: dashboardType,
				Args: graphql.FieldConfigArgument{"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					return svc.GetDashboard(context.Background(), uid)
				},
			},
		},
	})

	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"create_note": {
				Type: noteType,
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"title":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"content": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"tags":    &graphql.ArgumentConfig{Type: graphql.NewList(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					title, _ := p.Args["title"].(string)
					content, _ := p.Args["content"].(string)
					tagsRaw, _ := p.Args["tags"].([]any)
					tags := make([]string, 0, len(tagsRaw))
					for _, t := range tagsRaw {
						tags = append(tags, strings.TrimSpace(fmt.Sprintf("%v", t)))
					}
					return svc.CreateNote(context.Background(), uid, title, content, tags)
				},
			},
			"create_flashcard": {
				Type: flashcardType,
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"note_id": &graphql.ArgumentConfig{Type: graphql.String},
					"front":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"back":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					noteID, _ := p.Args["note_id"].(string)
					front, _ := p.Args["front"].(string)
					back, _ := p.Args["back"].(string)
					return svc.CreateFlashcard(context.Background(), uid, noteID, front, back)
				},
			},
			"generate_flashcards_from_note": {
				Type: graphql.NewList(flashcardType),
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"note_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"count":   &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					nid, _ := p.Args["note_id"].(string)
					count, _ := p.Args["count"].(int)
					return svc.GenerateFlashcardsFromNote(context.Background(), uid, nid, count)
				},
			},
			"review_flashcard": {
				Type: flashcardType,
				Args: graphql.FieldConfigArgument{
					"flashcard_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"quality":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					fid, _ := p.Args["flashcard_id"].(string)
					quality, _ := p.Args["quality"].(int)
					return svc.ReviewFlashcard(context.Background(), fid, quality)
				},
			},
			"create_quiz_from_note": {
				Type: quizType,
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"note_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"count":   &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					nid, _ := p.Args["note_id"].(string)
					count, _ := p.Args["count"].(int)
					return svc.CreateQuizFromNote(context.Background(), uid, nid, count)
				},
			},
			"submit_quiz_attempt": {
				Type: quizAttemptType,
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"quiz_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"answers": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					qid, _ := p.Args["quiz_id"].(string)
					pairs, _ := p.Args["answers"].([]any)
					answerMap := map[string]string{}
					for _, pair := range pairs {
						parts := strings.SplitN(fmt.Sprintf("%v", pair), "::", 2)
						if len(parts) == 2 {
							answerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
						}
					}
					return svc.SubmitQuizAttempt(context.Background(), uid, qid, answerMap)
				},
			},
			"save_study_session": {
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "StudySession",
					Fields: graphql.Fields{
						"id":         &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
						"duration":   &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
						"type":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
						"topic":      &graphql.Field{Type: graphql.String},
						"created_at": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("CreatedAt")},
					},
				}),
				Args: graphql.FieldConfigArgument{
					"user_id":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"duration": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"type":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"topic":    &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					uid, _ := p.Args["user_id"].(string)
					duration, _ := p.Args["duration"].(int)
					typ, _ := p.Args["type"].(string)
					topic, _ := p.Args["topic"].(string)
					return svc.SaveStudySession(context.Background(), uid, typ, topic, duration)
				},
			},
			"generate_mnemonic": {
				Type: graphql.NewNonNull(graphql.String),
				Args: graphql.FieldConfigArgument{
					"acronym": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"context": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					acr, _ := p.Args["acronym"].(string)
					ctxText, _ := p.Args["context"].(string)
					return svc.GenerateMnemonic(acr, ctxText), nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
}
