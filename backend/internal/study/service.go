package study

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"smart-study-assist-api/internal/database"
	"smart-study-assist-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	notesCol      *mongo.Collection
	flashcardsCol *mongo.Collection
	quizzesCol    *mongo.Collection
	sessionsCol   *mongo.Collection
}

type AcronymSuggestion struct {
	Acronym          string   `json:"acronym"`
	Score            float64  `json:"score"`
	ReadabilityScore float64  `json:"readability_score"`
	FamiliarityScore float64  `json:"familiarity_score"`
	Mnemonic         string   `json:"mnemonic"`
	SourceWords      []string `json:"source_words"`
}

type SmartNotesResult struct {
	Summary       string   `json:"summary"`
	KeyTerms      []string `json:"key_terms"`
	BulletPoints  []string `json:"bullet_points"`
	ExamReadyText string   `json:"exam_ready_text"`
}

type Dashboard struct {
	TotalStudySeconds int      `json:"total_study_seconds"`
	QuizAttempts      int      `json:"quiz_attempts"`
	AverageQuizScore  float64  `json:"average_quiz_score"`
	TopicsCovered     []string `json:"topics_covered"`
	WeakAreas         []string `json:"weak_areas"`
}

func NewService(client *mongo.Client) *Service {
	return &Service{
		notesCol:      database.GetCollection(client, "notes"),
		flashcardsCol: database.GetCollection(client, "flashcards"),
		quizzesCol:    database.GetCollection(client, "quizzes"),
		sessionsCol:   database.GetCollection(client, "study_sessions"),
	}
}

func (s *Service) GenerateAcronyms(words []string) []AcronymSuggestion {
	tokens := normalizeWords(words)
	if len(tokens) == 0 {
		return []AcronymSuggestion{}
	}

	letters := make([]string, 0, len(tokens))
	for _, token := range tokens {
		letters = append(letters, strings.ToUpper(string([]rune(token)[0])))
	}

	permutations := uniquePermutations(letters, 120)
	results := make([]AcronymSuggestion, 0, len(permutations))
	for _, perm := range permutations {
		acr := strings.Join(perm, "")
		readability := pronounceability(acr)
		familiarity := familiarityScore(acr)
		score := readability*0.65 + familiarity*0.35
		results = append(results, AcronymSuggestion{
			Acronym:          acr,
			Score:            round(score),
			ReadabilityScore: round(readability),
			FamiliarityScore: round(familiarity),
			Mnemonic:         s.GenerateMnemonic(acr, strings.Join(tokens, ", ")),
			SourceWords:      tokens,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > 20 {
		return results[:20]
	}
	return results
}

func (s *Service) GenerateMnemonic(acronym, contextText string) string {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey != "" {
		if phrase, err := generateMnemonicWithAI(apiKey, acronym, contextText); err == nil && strings.TrimSpace(phrase) != "" {
			return phrase
		}
	}

	wordbank := map[rune][]string{
		'A': {"Amazing", "Agile", "Active", "Accurate"},
		'B': {"Brave", "Bright", "Balanced", "Bold"},
		'C': {"Curious", "Calm", "Creative", "Clear"},
		'D': {"Dynamic", "Deep", "Direct", "Driven"},
		'E': {"Efficient", "Eager", "Elegant", "Epic"},
		'F': {"Focused", "Fast", "Friendly", "Flexible"},
		'G': {"Great", "Guided", "Generous", "Growing"},
		'H': {"Helpful", "Honest", "Healthy", "Humble"},
		'I': {"Incredible", "Insightful", "Intelligent", "Intentional"},
		'J': {"Joyful", "Just", "Jolly", "Jumping"},
		'K': {"Kind", "Keen", "Known", "Key"},
		'L': {"Logical", "Lively", "Light", "Learning"},
		'M': {"Mindful", "Modern", "Meaningful", "Motivated"},
		'N': {"Nimble", "Neat", "Natural", "Notable"},
		'O': {"Open", "Optimized", "Organized", "Outstanding"},
		'P': {"Precise", "Practical", "Positive", "Prepared"},
		'Q': {"Quick", "Quiet", "Qualified", "Questing"},
		'R': {"Reliable", "Rapid", "Ready", "Resilient"},
		'S': {"Smart", "Steady", "Skilled", "Sharp"},
		'T': {"Thoughtful", "Tidy", "Timely", "Trusted"},
		'U': {"Unique", "Unified", "Useful", "Upbeat"},
		'V': {"Vivid", "Valuable", "Versatile", "Visionary"},
		'W': {"Wise", "Warm", "Winning", "Working"},
		'X': {"Xenial", "Xtra", "Xact", "Xpressive"},
		'Y': {"Young", "Yielding", "Yearning", "Yellow"},
		'Z': {"Zesty", "Zealous", "Zonal", "Zippy"},
	}

	parts := make([]string, 0, len(acronym))
	for i, ch := range strings.ToUpper(acronym) {
		list, ok := wordbank[ch]
		if !ok {
			parts = append(parts, strings.ToUpper(string(ch)))
			continue
		}
		parts = append(parts, list[i%len(list)])
	}
	return strings.Join(parts, " ")
}

func (s *Service) SmartNotes(raw string) SmartNotesResult {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return SmartNotesResult{}
	}

	sentences := splitSentences(normalized)
	summary := strings.Join(take(sentences, 3), " ")
	terms := extractKeyTerms(normalized, 8)
	bullets := toBullets(sentences, 7)
	exam := strings.Join([]string{
		"Quick Revision Summary:",
		"- " + strings.Join(terms, ", "),
		"- " + strings.Join(take(bullets, 5), "\n- "),
		"- Focus: definitions, examples, and one real-world application per key term.",
	}, "\n")

	return SmartNotesResult{
		Summary:       summary,
		KeyTerms:      terms,
		BulletPoints:  bullets,
		ExamReadyText: exam,
	}
}

func (s *Service) CreateNote(ctx context.Context, userID, title, content string, tags []string) (*models.Note, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	smart := s.SmartNotes(content)
	now := time.Now().UTC()
	note := &models.Note{
		UserID:        uid,
		Title:         strings.TrimSpace(title),
		Content:       strings.TrimSpace(content),
		Summary:       smart.Summary,
		KeyTerms:      smart.KeyTerms,
		BulletPoints:  smart.BulletPoints,
		ExamReadyText: smart.ExamReadyText,
		Tags:          tags,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	res, err := s.notesCol.InsertOne(ctx, note)
	if err != nil {
		return nil, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		note.ID = oid
	}
	return note, nil
}

func (s *Service) ListNotes(ctx context.Context, userID string) ([]models.Note, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	cursor, err := s.notesCol.Find(ctx, bson.M{"user_id": uid}, options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := make([]models.Note, 0)
	for cursor.Next(ctx) {
		var n models.Note
		if err := cursor.Decode(&n); err != nil {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

func (s *Service) CreateFlashcard(ctx context.Context, userID, noteID, front, back string) (*models.Flashcard, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	card := &models.Flashcard{
		UserID:     uid,
		Front:      strings.TrimSpace(front),
		Back:       strings.TrimSpace(back),
		Interval:   1,
		EaseFactor: 2.5,
		Repetition: 0,
		NextReview: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}
	if strings.TrimSpace(noteID) != "" {
		if oid, err := primitive.ObjectIDFromHex(noteID); err == nil {
			card.NoteID = oid
		}
	}

	res, err := s.flashcardsCol.InsertOne(ctx, card)
	if err != nil {
		return nil, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		card.ID = oid
	}
	return card, nil
}

func (s *Service) GenerateFlashcardsFromNote(ctx context.Context, userID, noteID string, count int) ([]models.Flashcard, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	nid, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id")
	}

	var note models.Note
	if err := s.notesCol.FindOne(ctx, bson.M{"_id": nid, "user_id": uid}).Decode(&note); err != nil {
		return nil, err
	}

	if count <= 0 {
		count = 5
	}
	terms := take(note.KeyTerms, count)
	if len(terms) == 0 {
		terms = take(extractKeyTerms(note.Content, count), count)
	}

	cards := make([]models.Flashcard, 0, len(terms))
	for _, term := range terms {
		back := fmt.Sprintf("%s: derived from note '%s'.", term, note.Title)
		card, err := s.CreateFlashcard(ctx, userID, noteID, fmt.Sprintf("Explain %s", term), back)
		if err != nil {
			continue
		}
		cards = append(cards, *card)
	}
	return cards, nil
}

func (s *Service) ListFlashcards(ctx context.Context, userID string, dueOnly bool) ([]models.Flashcard, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	filter := bson.M{"user_id": uid}
	if dueOnly {
		filter["next_review"] = bson.M{"$lte": time.Now().UTC()}
	}
	cursor, err := s.flashcardsCol.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "next_review", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := make([]models.Flashcard, 0)
	for cursor.Next(ctx) {
		var card models.Flashcard
		if err := cursor.Decode(&card); err != nil {
			continue
		}
		out = append(out, card)
	}
	return out, nil
}

func (s *Service) ReviewFlashcard(ctx context.Context, flashcardID string, quality int) (*models.Flashcard, error) {
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}
	oid, err := primitive.ObjectIDFromHex(flashcardID)
	if err != nil {
		return nil, fmt.Errorf("invalid flashcard id")
	}
	var card models.Flashcard
	if err := s.flashcardsCol.FindOne(ctx, bson.M{"_id": oid}).Decode(&card); err != nil {
		return nil, err
	}

	if quality < 3 {
		card.Repetition = 0
		card.Interval = 1
	} else {
		card.Repetition++
		if card.Repetition == 1 {
			card.Interval = 1
		} else if card.Repetition == 2 {
			card.Interval = 6
		} else {
			card.Interval = int(math.Round(float64(card.Interval) * card.EaseFactor))
		}
	}

	newEF := card.EaseFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if newEF < 1.3 {
		newEF = 1.3
	}
	card.EaseFactor = newEF
	card.NextReview = time.Now().UTC().AddDate(0, 0, card.Interval)

	_, err = s.flashcardsCol.UpdateByID(ctx, oid, bson.M{"$set": bson.M{
		"interval":    card.Interval,
		"ease_factor": card.EaseFactor,
		"repetition":  card.Repetition,
		"next_review": card.NextReview,
	}})
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (s *Service) CreateQuizFromNote(ctx context.Context, userID, noteID string, count int) (*models.Quiz, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	nid, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id")
	}
	var note models.Note
	if err := s.notesCol.FindOne(ctx, bson.M{"_id": nid, "user_id": uid}).Decode(&note); err != nil {
		return nil, err
	}

	if count <= 0 {
		count = 6
	}
	terms := take(note.KeyTerms, count)
	if len(terms) == 0 {
		terms = extractKeyTerms(note.Content, count)
	}

	questions := make([]models.Question, 0, len(terms))
	for i, term := range terms {
		qid := primitive.NewObjectID().Hex()
		switch i % 3 {
		case 0:
			questions = append(questions, models.Question{
				ID:            qid,
				Text:          fmt.Sprintf("What best describes %s?", term),
				Type:          "mcq",
				Options:       []string{term + " is a key concept", term + " is unrelated", term + " is a date", term + " is a location"},
				CorrectAnswer: term + " is a key concept",
				Topic:         term,
			})
		case 1:
			questions = append(questions, models.Question{
				ID:            qid,
				Text:          fmt.Sprintf("True or False: %s appears as a key term in the note.", term),
				Type:          "true_false",
				Options:       []string{"True", "False"},
				CorrectAnswer: "True",
				Topic:         term,
			})
		default:
			questions = append(questions, models.Question{
				ID:            qid,
				Text:          fmt.Sprintf("Fill in the blank: ____ is an important idea in this note (%s).", note.Title),
				Type:          "fill_blank",
				CorrectAnswer: term,
				Topic:         term,
			})
		}
	}

	quiz := &models.Quiz{
		UserID:    uid,
		Title:     fmt.Sprintf("Quiz: %s", note.Title),
		Questions: questions,
		NoteID:    nid,
		CreatedAt: time.Now().UTC(),
	}

	res, err := s.quizzesCol.InsertOne(ctx, quiz)
	if err != nil {
		return nil, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		quiz.ID = oid
	}
	return quiz, nil
}

func (s *Service) ListQuizzes(ctx context.Context, userID string) ([]models.Quiz, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	cursor, err := s.quizzesCol.Find(ctx, bson.M{"user_id": uid}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := make([]models.Quiz, 0)
	for cursor.Next(ctx) {
		var quiz models.Quiz
		if err := cursor.Decode(&quiz); err != nil {
			continue
		}
		out = append(out, quiz)
	}
	return out, nil
}

func (s *Service) SubmitQuizAttempt(ctx context.Context, userID, quizID string, answers map[string]string) (*models.QuizAttempt, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	qid, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		return nil, fmt.Errorf("invalid quiz id")
	}
	var quiz models.Quiz
	if err := s.quizzesCol.FindOne(ctx, bson.M{"_id": qid, "user_id": uid}).Decode(&quiz); err != nil {
		return nil, err
	}

	correct := 0
	incorrectTopics := make([]string, 0)
	for _, q := range quiz.Questions {
		answer := strings.TrimSpace(strings.ToLower(answers[q.ID]))
		expected := strings.TrimSpace(strings.ToLower(q.CorrectAnswer))
		if answer == expected {
			correct++
		} else {
			incorrectTopics = append(incorrectTopics, q.Topic)
		}
	}

	score := 0.0
	if len(quiz.Questions) > 0 {
		score = (float64(correct) / float64(len(quiz.Questions))) * 100
	}

	attempt := &models.QuizAttempt{
		UserID:          uid,
		QuizID:          qid,
		Score:           round(score),
		IncorrectTopics: unique(incorrectTopics),
		CreatedAt:       time.Now().UTC(),
	}

	if _, err := s.quizzesCol.Database().Collection("quiz_attempts").InsertOne(ctx, attempt); err != nil {
		return nil, err
	}
	return attempt, nil
}

func (s *Service) SaveStudySession(ctx context.Context, userID, sessionType, topic string, duration int) (*models.StudySession, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	if duration < 1 {
		duration = 1
	}
	session := &models.StudySession{
		UserID:    uid,
		Duration:  duration,
		Type:      strings.TrimSpace(sessionType),
		Topic:     strings.TrimSpace(topic),
		CreatedAt: time.Now().UTC(),
	}
	res, err := s.sessionsCol.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		session.ID = oid
	}
	return session, nil
}

func (s *Service) GetDashboard(ctx context.Context, userID string) (*Dashboard, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	totalStudySeconds, _ := s.totalStudyTime(ctx, uid)
	quizAttempts, avgScore, weakAreas, _ := s.quizStats(ctx, uid)
	topics, _ := s.topicsCovered(ctx, uid)

	return &Dashboard{
		TotalStudySeconds: totalStudySeconds,
		QuizAttempts:      quizAttempts,
		AverageQuizScore:  round(avgScore),
		TopicsCovered:     topics,
		WeakAreas:         weakAreas,
	}, nil
}

func (s *Service) totalStudyTime(ctx context.Context, uid primitive.ObjectID) (int, error) {
	pipeline := mongo.Pipeline{{{Key: "$match", Value: bson.M{"user_id": uid}}}, {{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$duration"}}}}}
	cursor, err := s.sessionsCol.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	var row []bson.M
	if err := cursor.All(ctx, &row); err != nil || len(row) == 0 {
		return 0, err
	}
	switch v := row[0]["total"].(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, nil
	}
}

func (s *Service) quizStats(ctx context.Context, uid primitive.ObjectID) (int, float64, []string, error) {
	attemptsCol := s.quizzesCol.Database().Collection("quiz_attempts")
	cursor, err := attemptsCol.Find(ctx, bson.M{"user_id": uid})
	if err != nil {
		return 0, 0, nil, err
	}
	defer cursor.Close(ctx)

	count := 0
	scoreSum := 0.0
	weakCounter := map[string]int{}
	for cursor.Next(ctx) {
		var attempt models.QuizAttempt
		if err := cursor.Decode(&attempt); err != nil {
			continue
		}
		count++
		scoreSum += attempt.Score
		for _, t := range attempt.IncorrectTopics {
			weakCounter[t]++
		}
	}
	avg := 0.0
	if count > 0 {
		avg = scoreSum / float64(count)
	}

	type topicCount struct {
		Topic string
		Count int
	}
	pairs := make([]topicCount, 0, len(weakCounter))
	for k, v := range weakCounter {
		pairs = append(pairs, topicCount{Topic: k, Count: v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Count > pairs[j].Count })
	weak := make([]string, 0, 5)
	for _, p := range takeTopicCounts(pairs, 5) {
		weak = append(weak, p.Topic)
	}
	return count, avg, weak, nil
}

func (s *Service) topicsCovered(ctx context.Context, uid primitive.ObjectID) ([]string, error) {
	cursor, err := s.notesCol.Find(ctx, bson.M{"user_id": uid}, options.Find().SetProjection(bson.M{"tags": 1, "key_terms": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	seen := map[string]bool{}
	for cursor.Next(ctx) {
		var n models.Note
		if err := cursor.Decode(&n); err != nil {
			continue
		}
		for _, t := range append(n.Tags, n.KeyTerms...) {
			t = strings.TrimSpace(t)
			if t != "" {
				seen[t] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	if len(out) > 20 {
		return out[:20], nil
	}
	return out, nil
}

func normalizeWords(words []string) []string {
	clean := make([]string, 0, len(words))
	re := regexp.MustCompile(`[A-Za-z]+`)
	for _, w := range words {
		for _, token := range re.FindAllString(w, -1) {
			t := strings.TrimSpace(token)
			if t != "" {
				clean = append(clean, strings.ToLower(t))
			}
		}
	}
	return clean
}

func uniquePermutations(chars []string, limit int) [][]string {
	if len(chars) == 0 {
		return [][]string{}
	}
	if len(chars) > 7 {
		chars = chars[:7]
	}
	sort.Strings(chars)
	used := make([]bool, len(chars))
	current := make([]string, 0, len(chars))
	results := make([][]string, 0, limit)

	var backtrack func()
	backtrack = func() {
		if len(results) >= limit {
			return
		}
		if len(current) == len(chars) {
			perm := make([]string, len(current))
			copy(perm, current)
			results = append(results, perm)
			return
		}
		for i := 0; i < len(chars); i++ {
			if used[i] {
				continue
			}
			if i > 0 && chars[i] == chars[i-1] && !used[i-1] {
				continue
			}
			used[i] = true
			current = append(current, chars[i])
			backtrack()
			current = current[:len(current)-1]
			used[i] = false
		}
	}
	backtrack()
	return results
}

func pronounceability(acr string) float64 {
	acr = strings.ToUpper(acr)
	if acr == "" {
		return 0
	}
	vowels := "AEIOU"
	vowelCount := 0
	runs := 0
	prevVowel := false
	for i, ch := range acr {
		isVowel := strings.ContainsRune(vowels, ch)
		if isVowel {
			vowelCount++
		}
		if i == 0 || isVowel != prevVowel {
			runs++
		}
		prevVowel = isVowel
	}
	ratio := float64(vowelCount) / float64(len(acr))
	balance := 1 - math.Abs(0.4-ratio)
	alternation := 1 / float64(runs)
	score := (balance * 0.8) + ((1-alternation) * 0.2)
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func familiarityScore(acr string) float64 {
	patterns := []string{"ST", "PR", "TR", "CH", "PH", "GR", "CL", "BR", "TH", "FL"}
	acr = strings.ToUpper(acr)
	if len(acr) < 2 {
		return 0.2
	}
	hits := 0
	for _, p := range patterns {
		if strings.Contains(acr, p) {
			hits++
		}
	}
	return math.Min(1.0, float64(hits+1)/float64(len(acr)))
}

func splitSentences(raw string) []string {
	re := regexp.MustCompile(`[.!?]\s+|\n+`)
	parts := re.Split(raw, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{strings.TrimSpace(raw)}
	}
	return out
}

func extractKeyTerms(raw string, limit int) []string {
	stop := map[string]bool{
		"the": true, "and": true, "for": true, "that": true, "with": true, "from": true, "this": true,
		"into": true, "have": true, "has": true, "were": true, "they": true, "their": true, "about": true,
		"your": true, "will": true, "would": true, "should": true, "can": true, "could": true, "is": true,
		"are": true, "in": true, "of": true, "to": true, "on": true, "a": true, "an": true,
	}
	terms := map[string]int{}
	re := regexp.MustCompile(`[A-Za-z]{3,}`)
	for _, word := range re.FindAllString(strings.ToLower(raw), -1) {
		if stop[word] {
			continue
		}
		terms[word]++
	}
	type kv struct {
		Word  string
		Count int
	}
	list := make([]kv, 0, len(terms))
	for k, v := range terms {
		list = append(list, kv{Word: k, Count: v})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].Word < list[j].Word
		}
		return list[i].Count > list[j].Count
	})
	if limit <= 0 {
		limit = 8
	}
	if len(list) > limit {
		list = list[:limit]
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		out = append(out, item.Word)
	}
	return out
}

func toBullets(sentences []string, limit int) []string {
	if limit <= 0 {
		limit = 6
	}
	if len(sentences) > limit {
		sentences = sentences[:limit]
	}
	out := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.HasSuffix(s, ".") {
			s += "."
		}
		out = append(out, s)
	}
	return out
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

func take[T any](list []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if len(list) <= n {
		return list
	}
	return list[:n]
}

func takeTopicCounts(list []struct {
	Topic string
	Count int
}, n int) []struct {
	Topic string
	Count int
} {
	if len(list) <= n {
		return list
	}
	return list[:n]
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func generateMnemonicWithAI(apiKey, acronym, contextText string) (string, error) {
	body := map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "Generate one concise mnemonic sentence from an acronym. Keep it natural and memorable."},
			{"role": "user", "content": "Acronym: " + acronym + "\nContext: " + contextText},
		},
		"temperature": 0.7,
	}
	buf, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", strings.NewReader(string(buf)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai request failed: %s", strconv.Itoa(resp.StatusCode))
	}

	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Choices) == 0 {
		return "", fmt.Errorf("no choices")
	}
	return strings.TrimSpace(payload.Choices[0].Message.Content), nil
}
