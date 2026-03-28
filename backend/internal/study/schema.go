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
			"acronym": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"score":   &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"readabilityScore": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if src, ok := p.Source.(AcronymSuggestion); ok {
						return src.ReadabilityScore, nil
					}
					if src, ok := p.Source.(*AcronymSuggestion); ok {
						return src.ReadabilityScore, nil
					}
					return 0.0, nil
				},
			},
			"familiarityScore": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if src, ok := p.Source.(AcronymSuggestion); ok {
						return src.FamiliarityScore, nil
					}
					if src, ok := p.Source.(*AcronymSuggestion); ok {
						return src.FamiliarityScore, nil
					}
					return 0.0, nil
				},
			},
			"mnemonic": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"sourceWords": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if src, ok := p.Source.(AcronymSuggestion); ok {
						return src.SourceWords, nil
					}
					if src, ok := p.Source.(*AcronymSuggestion); ok {
						return src.SourceWords, nil
					}
					return []string{}, nil
				},
			},
		},
	})

	smartNotesType := graphql.NewObject(graphql.ObjectConfig{
		Name: "SmartNotesResult",
		Fields: graphql.Fields{
			"summary": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"keyTerms": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if src, ok := p.Source.(SmartNotesResult); ok {
						return src.KeyTerms, nil
					}
					if src, ok := p.Source.(*SmartNotesResult); ok {
						return src.KeyTerms, nil
					}
					return []string{}, nil
				},
			},
			"bulletPoints": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if src, ok := p.Source.(SmartNotesResult); ok {
						return src.BulletPoints, nil
					}
					if src, ok := p.Source.(*SmartNotesResult); ok {
						return src.BulletPoints, nil
					}
					return []string{}, nil
				},
			},
			"examReadyText": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if src, ok := p.Source.(SmartNotesResult); ok {
						return src.ExamReadyText, nil
					}
					if src, ok := p.Source.(*SmartNotesResult); ok {
						return src.ExamReadyText, nil
					}
					return "", nil
				},
			},
		},
	})

	questionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Question",
		Fields: graphql.Fields{
			"id":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"text":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"type":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"options": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"correctAnswer": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return structStringField(p.Source, "CorrectAnswer"), nil
				},
			},
			"topic": &graphql.Field{Type: graphql.String},
		},
	})

	noteType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Note",
		Fields: graphql.Fields{
			"id":            &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"title":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"content":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"summary":       &graphql.Field{Type: graphql.String},
			"keyTerms":      &graphql.Field{Type: graphql.NewList(graphql.String), Resolve: stringSliceField("KeyTerms")},
			"bulletPoints":  &graphql.Field{Type: graphql.NewList(graphql.String), Resolve: stringSliceField("BulletPoints")},
			"examReadyText": &graphql.Field{Type: graphql.String, Resolve: stringField("ExamReadyText")},
			"tags":          &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	flashcardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Flashcard",
		Fields: graphql.Fields{
			"id":         &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"noteId":     &graphql.Field{Type: graphql.String, Resolve: oidResolver("NoteID")},
			"front":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"back":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"nextReview": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("NextReview")},
			"interval":   &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"easeFactor": &graphql.Field{Type: graphql.NewNonNull(graphql.Float), Resolve: floatField("EaseFactor")},
			"repetition": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})

	quizType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Quiz",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"title":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"questions": &graphql.Field{Type: graphql.NewList(questionType)},
			"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("CreatedAt")},
		},
	})

	quizAttemptType := graphql.NewObject(graphql.ObjectConfig{
		Name: "QuizAttempt",
		Fields: graphql.Fields{
			"id":              &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"score":           &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"incorrectTopics": &graphql.Field{Type: graphql.NewList(graphql.String), Resolve: stringSliceField("IncorrectTopics")},
			"createdAt":       &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("CreatedAt")},
		},
	})

	studySessionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "StudySession",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: oidResolver("ID")},
			"duration":  &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"type":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"topic":     &graphql.Field{Type: graphql.String},
			"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: timeResolver("CreatedAt")},
		},
	})

	dashboardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Dashboard",
		Fields: graphql.Fields{
			"totalStudySeconds": &graphql.Field{Type: graphql.NewNonNull(graphql.Int), Resolve: intField("TotalStudySeconds")},
			"quizAttempts":      &graphql.Field{Type: graphql.NewNonNull(graphql.Int), Resolve: intField("QuizAttempts")},
			"averageQuizScore":  &graphql.Field{Type: graphql.NewNonNull(graphql.Float), Resolve: floatField("AverageQuizScore")},
			"topicsCovered":     &graphql.Field{Type: graphql.NewList(graphql.String), Resolve: stringSliceField("TopicsCovered")},
			"weakAreas":         &graphql.Field{Type: graphql.NewList(graphql.String), Resolve: stringSliceField("WeakAreas")},
		},
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{Type: graphql.NewNonNull(graphql.String), Resolve: func(p graphql.ResolveParams) (interface{}, error) { return "ok", nil }},
			"generateAcronyms": &graphql.Field{
				Type: graphql.NewList(acronymType),
				Args: graphql.FieldConfigArgument{
					"words": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					wordsAny, _ := p.Args["words"].([]interface{})
					words := make([]string, 0, len(wordsAny))
					for _, v := range wordsAny {
						words = append(words, fmt.Sprintf("%v", v))
					}
					return svc.GenerateAcronyms(words), nil
				},
			},
			"smartNotes": &graphql.Field{
				Type: smartNotesType,
				Args: graphql.FieldConfigArgument{
					"raw": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					raw, _ := p.Args["raw"].(string)
					return svc.SmartNotes(raw), nil
				},
			},
			"notes": &graphql.Field{
				Type: graphql.NewList(noteType),
				Args: graphql.FieldConfigArgument{"userId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					return svc.ListNotes(context.Background(), userID)
				},
			},
			"flashcards": &graphql.Field{
				Type: graphql.NewList(flashcardType),
				Args: graphql.FieldConfigArgument{
					"userId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"dueOnly": &graphql.ArgumentConfig{Type: graphql.Boolean},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					dueOnly, _ := p.Args["dueOnly"].(bool)
					return svc.ListFlashcards(context.Background(), userID, dueOnly)
				},
			},
			"quizzes": &graphql.Field{
				Type: graphql.NewList(quizType),
				Args: graphql.FieldConfigArgument{"userId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					return svc.ListQuizzes(context.Background(), userID)
				},
			},
			"dashboard": &graphql.Field{
				Type: dashboardType,
				Args: graphql.FieldConfigArgument{"userId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					return svc.GetDashboard(context.Background(), userID)
				},
			},
		},
	})

	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createNote": &graphql.Field{
				Type: noteType,
				Args: graphql.FieldConfigArgument{
					"userId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"title":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"content": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"tags":    &graphql.ArgumentConfig{Type: graphql.NewList(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					title, _ := p.Args["title"].(string)
					content, _ := p.Args["content"].(string)
					tagsAny, _ := p.Args["tags"].([]interface{})
					tags := make([]string, 0, len(tagsAny))
					for _, t := range tagsAny {
						tags = append(tags, strings.TrimSpace(fmt.Sprintf("%v", t)))
					}
					return svc.CreateNote(context.Background(), userID, title, content, tags)
				},
			},
			"createFlashcard": &graphql.Field{
				Type: flashcardType,
				Args: graphql.FieldConfigArgument{
					"userId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"noteId": &graphql.ArgumentConfig{Type: graphql.String},
					"front":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"back":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					noteID, _ := p.Args["noteId"].(string)
					front, _ := p.Args["front"].(string)
					back, _ := p.Args["back"].(string)
					return svc.CreateFlashcard(context.Background(), userID, noteID, front, back)
				},
			},
			"generateFlashcardsFromNote": &graphql.Field{
				Type: graphql.NewList(flashcardType),
				Args: graphql.FieldConfigArgument{
					"userId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"noteId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"count":  &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					noteID, _ := p.Args["noteId"].(string)
					count, _ := p.Args["count"].(int)
					return svc.GenerateFlashcardsFromNote(context.Background(), userID, noteID, count)
				},
			},
			"reviewFlashcard": &graphql.Field{
				Type: flashcardType,
				Args: graphql.FieldConfigArgument{
					"flashcardId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"quality":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					flashcardID, _ := p.Args["flashcardId"].(string)
					quality, _ := p.Args["quality"].(int)
					return svc.ReviewFlashcard(context.Background(), flashcardID, quality)
				},
			},
			"createQuizFromNote": &graphql.Field{
				Type: quizType,
				Args: graphql.FieldConfigArgument{
					"userId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"noteId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"count":  &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					noteID, _ := p.Args["noteId"].(string)
					count, _ := p.Args["count"].(int)
					return svc.CreateQuizFromNote(context.Background(), userID, noteID, count)
				},
			},
			"submitQuizAttempt": &graphql.Field{
				Type: quizAttemptType,
				Args: graphql.FieldConfigArgument{
					"userId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"quizId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"answers": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					quizID, _ := p.Args["quizId"].(string)
					pairs, _ := p.Args["answers"].([]interface{})
					answers := map[string]string{}
					for _, pair := range pairs {
						parts := strings.SplitN(fmt.Sprintf("%v", pair), "::", 2)
						if len(parts) == 2 {
							answers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
						}
					}
					return svc.SubmitQuizAttempt(context.Background(), userID, quizID, answers)
				},
			},
			"saveStudySession": &graphql.Field{
				Type: studySessionType,
				Args: graphql.FieldConfigArgument{
					"userId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"duration": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"type":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"topic":    &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["userId"].(string)
					duration, _ := p.Args["duration"].(int)
					sessionType, _ := p.Args["type"].(string)
					topic, _ := p.Args["topic"].(string)
					return svc.SaveStudySession(context.Background(), userID, sessionType, topic, duration)
				},
			},
			"generateMnemonic": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Args: graphql.FieldConfigArgument{
					"acronym": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"context": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acronym, _ := p.Args["acronym"].(string)
					contextText, _ := p.Args["context"].(string)
					return svc.GenerateMnemonic(acronym, contextText), nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
}
