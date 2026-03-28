'use client';

import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';

type View =
  | 'command'
  | 'notes'
  | 'mnemonics'
  | 'flashcards'
  | 'quizzes'
  | 'focus';

type Note = {
  id: string;
  title: string;
  content: string;
  summary?: string;
  keyTerms?: string[];
  bulletPoints?: string[];
  examReadyText?: string;
  tags?: string[];
};

type AcronymSuggestion = {
  acronym: string;
  score: number;
  readabilityScore: number;
  familiarityScore: number;
  mnemonic: string;
  sourceWords: string[];
};

type Flashcard = {
  id: string;
  noteId?: string;
  front: string;
  back: string;
  nextReview: string;
  interval: number;
  easeFactor: number;
  repetition: number;
};

type QuizQuestion = {
  id: string;
  text: string;
  type: string;
  options?: string[];
  correctAnswer: string;
  topic?: string;
};

type Quiz = {
  id: string;
  title: string;
  createdAt: string;
  questions: QuizQuestion[];
};

type Dashboard = {
  totalStudySeconds: number;
  quizAttempts: number;
  averageQuizScore: number;
  topicsCovered: string[];
  weakAreas: string[];
};

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/graphql';
const DEFAULT_USER_ID = '64b64f7c5e9b1f0d0a1b2c3d';
const DAILY_GOAL_STORAGE_KEY = 'ssa-daily-goal-minutes';
const TODAY_FOCUS_STORAGE_KEY = 'ssa-today-focus-seconds';

async function gql<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const res = await fetch(API_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, variables }),
  });

  if (!res.ok) {
    throw new Error(`GraphQL request failed (${res.status})`);
  }

  const payload = await res.json();
  if (payload.errors?.length) {
    throw new Error(payload.errors[0]?.message ?? 'GraphQL error');
  }

  return payload.data as T;
}

function formatClock(secondsTotal: number): string {
  const minutes = Math.floor(secondsTotal / 60);
  const seconds = secondsTotal % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}

export default function Home() {
  const userId = DEFAULT_USER_ID;

  const [activeView, setActiveView] = useState<View>('command');
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState('Ready');

  const [notes, setNotes] = useState<Note[]>([]);
  const [flashcards, setFlashcards] = useState<Flashcard[]>([]);
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);

  const [globalSearch, setGlobalSearch] = useState('');
  const [dailyGoalMinutes, setDailyGoalMinutes] = useState(90);
  const [todayFocusSeconds, setTodayFocusSeconds] = useState(0);
  const [dueOnlyFlashcards, setDueOnlyFlashcards] = useState(false);

  const [noteTitle, setNoteTitle] = useState('');
  const [noteTags, setNoteTags] = useState('');
  const [noteContent, setNoteContent] = useState('');

  const [acronymWords, setAcronymWords] = useState('');
  const [acronyms, setAcronyms] = useState<AcronymSuggestion[]>([]);

  const [manualCardFront, setManualCardFront] = useState('');
  const [manualCardBack, setManualCardBack] = useState('');
  const [selectedNoteId, setSelectedNoteId] = useState('');
  const [quickReviewCard, setQuickReviewCard] = useState<Flashcard | null>(
    null,
  );
  const [showQuickReviewAnswer, setShowQuickReviewAnswer] = useState(false);

  const [activeQuizId, setActiveQuizId] = useState('');
  const [quizAnswers, setQuizAnswers] = useState<Record<string, string>>({});
  const [lastQuizScore, setLastQuizScore] = useState<number | null>(null);
  const [randomQuestion, setRandomQuestion] = useState<QuizQuestion | null>(
    null,
  );
  const [showRandomAnswer, setShowRandomAnswer] = useState(false);

  const [pomodoroOn, setPomodoroOn] = useState(false);
  const [secondsLeft, setSecondsLeft] = useState(25 * 60);
  const [mode, setMode] = useState<'focus' | 'break'>('focus');
  const [studyTopic, setStudyTopic] = useState('Deep Work');

  const navItems: { id: View; label: string; hint: string }[] = [
    { id: 'command', label: 'Command Center', hint: 'Overview + goals' },
    { id: 'notes', label: 'Notes Studio', hint: 'Create + search notes' },
    { id: 'mnemonics', label: 'Mnemonic Forge', hint: 'Acronym engine' },
    { id: 'flashcards', label: 'Flashcards', hint: 'Spaced repetition' },
    { id: 'quizzes', label: 'Quiz Arena', hint: 'Practice and score' },
    { id: 'focus', label: 'Focus Timer', hint: 'Pomodoro loop' },
  ];

  const activeQuiz = useMemo(
    () => quizzes.find((quiz) => quiz.id === activeQuizId),
    [quizzes, activeQuizId],
  );

  const searchNeedle = globalSearch.trim().toLowerCase();

  const filteredNotes = useMemo(() => {
    if (!searchNeedle) {
      return notes;
    }

    return notes.filter((note) => {
      const haystack = [
        note.title,
        note.content,
        note.summary,
        ...(note.keyTerms ?? []),
        ...(note.tags ?? []),
      ]
        .join(' ')
        .toLowerCase();

      return haystack.includes(searchNeedle);
    });
  }, [notes, searchNeedle]);

  const filteredFlashcards = useMemo(() => {
    if (!searchNeedle) {
      return flashcards;
    }

    return flashcards.filter((card) => {
      const haystack = `${card.front} ${card.back}`.toLowerCase();
      return haystack.includes(searchNeedle);
    });
  }, [flashcards, searchNeedle]);

  const filteredQuizzes = useMemo(() => {
    if (!searchNeedle) {
      return quizzes;
    }

    return quizzes.filter((quiz) => {
      const haystack = [
        quiz.title,
        ...quiz.questions.map(
          (question) => `${question.text} ${question.topic ?? ''}`,
        ),
      ]
        .join(' ')
        .toLowerCase();

      return haystack.includes(searchNeedle);
    });
  }, [quizzes, searchNeedle]);

  const goalProgress = Math.min(
    100,
    Math.round((todayFocusSeconds / Math.max(dailyGoalMinutes * 60, 1)) * 100),
  );

  const dueSoonCount = flashcards.filter((card) => {
    return new Date(card.nextReview).getTime() <= Date.now();
  }).length;

  const loadAll = useCallback(async () => {
    setLoading(true);
    try {
      const [noteData, flashcardData, quizData, dashboardData] =
        await Promise.all([
          gql<{ notes: Note[] }>(
            `query($userId: String!) { notes(userId: $userId) { id title content summary keyTerms bulletPoints examReadyText tags } }`,
            { userId },
          ),
          gql<{ flashcards: Flashcard[] }>(
            `query($userId: String!, $dueOnly: Boolean) { flashcards(userId: $userId, dueOnly: $dueOnly) { id noteId front back nextReview interval easeFactor repetition } }`,
            { userId, dueOnly: dueOnlyFlashcards },
          ),
          gql<{ quizzes: Quiz[] }>(
            `query($userId: String!) { quizzes(userId: $userId) { id title createdAt questions { id text type options correctAnswer topic } } }`,
            { userId },
          ),
          gql<{ dashboard: Dashboard }>(
            `query($userId: String!) { dashboard(userId: $userId) { totalStudySeconds quizAttempts averageQuizScore topicsCovered weakAreas } }`,
            { userId },
          ),
        ]);

      setNotes(noteData.notes ?? []);
      setFlashcards(flashcardData.flashcards ?? []);
      setQuizzes(quizData.quizzes ?? []);
      setDashboard(dashboardData.dashboard ?? null);
      setStatus('Synced with API');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  }, [dueOnlyFlashcards, userId]);

  const savePomodoroSession = useCallback(
    async (
      sessionType: 'focus' | 'break' = mode,
      durationSeconds = (sessionType === 'focus' ? 25 : 5) * 60,
    ) => {
      setLoading(true);
      try {
        await gql(
          `mutation($userId: String!, $duration: Int!, $type: String!, $topic: String) {
          saveStudySession(userId: $userId, duration: $duration, type: $type, topic: $topic) { id }
        }`,
          {
            userId,
            duration: durationSeconds,
            type: sessionType === 'focus' ? 'pomodoro' : 'break',
            topic: studyTopic,
          },
        );
        setStatus('Study session saved.');
        await loadAll();
      } catch (error) {
        setStatus(
          error instanceof Error
            ? error.message
            : 'Could not save study session',
        );
      } finally {
        setLoading(false);
      }
    },
    [loadAll, mode, studyTopic, userId],
  );

  useEffect(() => {
    const savedGoal = window.localStorage.getItem(DAILY_GOAL_STORAGE_KEY);
    const savedFocus = window.localStorage.getItem(TODAY_FOCUS_STORAGE_KEY);

    if (savedGoal) {
      const parsed = Number(savedGoal);
      if (!Number.isNaN(parsed) && parsed > 0) {
        setDailyGoalMinutes(parsed);
      }
    }

    if (savedFocus) {
      const parsed = Number(savedFocus);
      if (!Number.isNaN(parsed) && parsed >= 0) {
        setTodayFocusSeconds(parsed);
      }
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem(
      DAILY_GOAL_STORAGE_KEY,
      String(dailyGoalMinutes),
    );
  }, [dailyGoalMinutes]);

  useEffect(() => {
    window.localStorage.setItem(
      TODAY_FOCUS_STORAGE_KEY,
      String(todayFocusSeconds),
    );
  }, [todayFocusSeconds]);

  useEffect(() => {
    void loadAll();
  }, [loadAll]);

  useEffect(() => {
    if (!pomodoroOn) {
      return;
    }

    const id = window.setInterval(() => {
      setSecondsLeft((curr) => Math.max(curr - 1, 0));
    }, 1000);

    return () => window.clearInterval(id);
  }, [pomodoroOn]);

  useEffect(() => {
    if (!pomodoroOn || secondsLeft !== 0) {
      return;
    }

    setMode((prev) => {
      const completedMode = prev;
      const nextMode = prev === 'focus' ? 'break' : 'focus';
      const completedSeconds = completedMode === 'focus' ? 25 * 60 : 5 * 60;

      if (completedMode === 'focus') {
        setTodayFocusSeconds((curr) => curr + completedSeconds);
      }

      void savePomodoroSession(completedMode, completedSeconds);
      setSecondsLeft(nextMode === 'focus' ? 25 * 60 : 5 * 60);
      return nextMode;
    });
  }, [pomodoroOn, savePomodoroSession, secondsLeft]);

  async function onCreateNote(e: FormEvent) {
    e.preventDefault();
    if (!noteTitle.trim() || !noteContent.trim()) {
      setStatus('Title and content are required.');
      return;
    }

    setLoading(true);
    try {
      const tags = noteTags
        .split(',')
        .map((x) => x.trim())
        .filter(Boolean);

      await gql(
        `mutation($userId: String!, $title: String!, $content: String!, $tags: [String!]) {
          createNote(userId: $userId, title: $title, content: $content, tags: $tags) { id }
        }`,
        { userId, title: noteTitle, content: noteContent, tags },
      );

      setNoteTitle('');
      setNoteContent('');
      setNoteTags('');
      setStatus('Note created with smart summary + key terms.');
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Could not create note',
      );
    } finally {
      setLoading(false);
    }
  }

  async function onGenerateAcronyms(e: FormEvent) {
    e.preventDefault();
    const words = acronymWords
      .split(/\n|,|;/)
      .map((x) => x.trim())
      .filter(Boolean);

    if (!words.length) {
      setStatus('Please enter at least one word or bullet point.');
      return;
    }

    setLoading(true);
    try {
      const data = await gql<{ generateAcronyms: AcronymSuggestion[] }>(
        `query($words: [String!]!) {
          generateAcronyms(words: $words) { acronym score readabilityScore familiarityScore mnemonic sourceWords }
        }`,
        { words },
      );
      setAcronyms(data.generateAcronyms ?? []);
      setStatus('Acronym set generated and ranked.');
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Acronym generation failed',
      );
    } finally {
      setLoading(false);
    }
  }

  async function copyMnemonic(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      setStatus('Mnemonic copied to clipboard.');
    } catch {
      setStatus('Clipboard not available in this browser.');
    }
  }

  async function createManualCard() {
    if (!manualCardFront.trim() || !manualCardBack.trim()) {
      setStatus('Both front and back are required.');
      return;
    }

    setLoading(true);
    try {
      await gql(
        `mutation($userId: String!, $noteId: String, $front: String!, $back: String!) {
          createFlashcard(userId: $userId, noteId: $noteId, front: $front, back: $back) { id }
        }`,
        {
          userId,
          noteId: selectedNoteId || null,
          front: manualCardFront,
          back: manualCardBack,
        },
      );
      setManualCardFront('');
      setManualCardBack('');
      setStatus('Flashcard created.');
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Failed to create flashcard',
      );
    } finally {
      setLoading(false);
    }
  }

  async function onCreateManualCard(e: FormEvent) {
    e.preventDefault();
    await createManualCard();
  }

  function pickQuickReviewCard() {
    const now = Date.now();
    const dueCards = flashcards.filter(
      (card) => new Date(card.nextReview).getTime() <= now,
    );
    const source = dueCards.length ? dueCards : flashcards;

    if (!source.length) {
      setStatus('No flashcards available yet. Create one first.');
      return;
    }

    const randomIndex = Math.floor(Math.random() * source.length);
    setQuickReviewCard(source[randomIndex]);
    setShowQuickReviewAnswer(false);
    setStatus(dueCards.length ? 'Loaded a due card.' : 'Loaded a random card.');
  }

  async function generateCardsFromNote(noteId: string) {
    setLoading(true);
    try {
      await gql(
        `mutation($userId: String!, $noteId: String!, $count: Int) {
          generateFlashcardsFromNote(userId: $userId, noteId: $noteId, count: $count) { id }
        }`,
        { userId, noteId, count: 6 },
      );
      setStatus('Auto-generated flashcards from note.');
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error
          ? error.message
          : 'Failed to auto-generate cards',
      );
    } finally {
      setLoading(false);
    }
  }

  async function reviewCard(flashcardId: string, quality: number) {
    setLoading(true);
    try {
      await gql(
        `mutation($flashcardId: String!, $quality: Int!) {
          reviewFlashcard(flashcardId: $flashcardId, quality: $quality) { id interval nextReview repetition }
        }`,
        { flashcardId, quality },
      );
      setStatus(`Review saved (quality ${quality}).`);
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Could not save review',
      );
    } finally {
      setLoading(false);
    }
  }

  async function createQuizFromNote(noteId: string) {
    setLoading(true);
    try {
      await gql(
        `mutation($userId: String!, $noteId: String!, $count: Int) {
          createQuizFromNote(userId: $userId, noteId: $noteId, count: $count) { id }
        }`,
        { userId, noteId, count: 6 },
      );
      setStatus('Quiz generated from note.');
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Could not generate quiz',
      );
    } finally {
      setLoading(false);
    }
  }

  async function submitQuiz() {
    if (!activeQuiz) {
      return;
    }

    const answers = Object.entries(quizAnswers).map(
      ([questionId, answer]) => `${questionId}::${answer}`,
    );

    setLoading(true);
    try {
      const data = await gql<{ submitQuizAttempt: { score: number } }>(
        `mutation($userId: String!, $quizId: String!, $answers: [String!]!) {
          submitQuizAttempt(userId: $userId, quizId: $quizId, answers: $answers) { score }
        }`,
        { userId, quizId: activeQuiz.id, answers },
      );

      setLastQuizScore(data.submitQuizAttempt.score);
      setStatus('Quiz submitted. Weak areas are tracked in dashboard.');
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Could not submit quiz',
      );
    } finally {
      setLoading(false);
    }
  }

  function pickRandomQuestion() {
    if (!activeQuiz?.questions.length) {
      setStatus('Pick a quiz first to use random practice mode.');
      return;
    }

    const idx = Math.floor(Math.random() * activeQuiz.questions.length);
    setRandomQuestion(activeQuiz.questions[idx]);
    setShowRandomAnswer(false);
    setStatus('Random question ready. Try answering before reveal.');
  }

  function renderCommandCenter() {
    const totalHours = ((dashboard?.totalStudySeconds ?? 0) / 3600).toFixed(1);

    return (
      <section className="view-stack">
        <div className="panel panel-heavy hero-grid">
          <div>
            <p className="eyebrow">System status</p>
            <h2 className="hero-title">Noir Study Console</h2>
            <p className="muted">
              Rebuilt for focus: one search, one command surface, zero visible
              user IDs.
            </p>
          </div>
          <div className="metrics-row">
            <div className="metric-card">
              <p className="metric-label">Total Study</p>
              <p className="metric-value">{totalHours}h</p>
            </div>
            <div className="metric-card">
              <p className="metric-label">Quiz Avg</p>
              <p className="metric-value">
                {dashboard?.averageQuizScore ?? 0}%
              </p>
            </div>
            <div className="metric-card">
              <p className="metric-label">Due Cards</p>
              <p className="metric-value">{dueSoonCount}</p>
            </div>
          </div>
        </div>

        <div className="grid-duo">
          <div className="panel">
            <h3>Daily Goal Tracker</h3>
            <p className="muted small">
              Stored locally for quick momentum checks.
            </p>
            <div className="goal-input-row">
              <label htmlFor="goal">Goal minutes</label>
              <input
                id="goal"
                className="input"
                type="number"
                min={15}
                step={5}
                value={dailyGoalMinutes}
                onChange={(e) =>
                  setDailyGoalMinutes(
                    Math.max(15, Number(e.target.value) || 15),
                  )
                }
              />
            </div>
            <p className="muted small">
              Today focus: {Math.round(todayFocusSeconds / 60)} min
            </p>
            <div className="progress-shell">
              <div
                className="progress-bar"
                style={{ width: `${goalProgress}%` }}
              />
            </div>
            <p className="small">{goalProgress}% complete</p>
          </div>

          <div className="panel">
            <h3>Weak Areas</h3>
            <div className="tag-wrap">
              {(dashboard?.weakAreas ?? []).map((topic) => (
                <span key={topic} className="chip chip-warn">
                  {topic}
                </span>
              ))}
              {!dashboard?.weakAreas?.length && (
                <span className="muted small">No weak areas flagged yet.</span>
              )}
            </div>
            <h3 className="space-top">Topics Covered</h3>
            <div className="tag-wrap">
              {(dashboard?.topicsCovered ?? []).map((topic) => (
                <span key={topic} className="chip">
                  {topic}
                </span>
              ))}
              {!dashboard?.topicsCovered?.length && (
                <span className="muted small">No topics recorded.</span>
              )}
            </div>
          </div>

          <div className="panel panel-heavy form-stack">
            <div className="item-head">
              <h3>Quick Flashcard Creator</h3>
              <button
                className="button button-ghost"
                onClick={() => setActiveView('flashcards')}
              >
                Open Flashcards Tab
              </button>
            </div>
            <select
              className="select"
              value={selectedNoteId}
              onChange={(e) => setSelectedNoteId(e.target.value)}
            >
              <option value="">Optional note link</option>
              {notes.map((note) => (
                <option key={note.id} value={note.id}>
                  {note.title}
                </option>
              ))}
            </select>
            <input
              className="input"
              placeholder="Front"
              value={manualCardFront}
              onChange={(e) => setManualCardFront(e.target.value)}
            />
            <textarea
              className="textarea"
              placeholder="Back"
              value={manualCardBack}
              onChange={(e) => setManualCardBack(e.target.value)}
            />
            <button
              className="button button-primary"
              onClick={() => void createManualCard()}
            >
              Create Flashcard Now
            </button>
          </div>
        </div>

        <div className="grid-duo">
          <div className="panel form-stack">
            <h3>Note Boost Actions</h3>
            <p className="muted small">
              Pick a note and instantly generate quiz and flashcard packs.
            </p>
            <select
              className="select"
              value={selectedNoteId}
              onChange={(e) => setSelectedNoteId(e.target.value)}
            >
              <option value="">Choose note</option>
              {notes.map((note) => (
                <option key={note.id} value={note.id}>
                  {note.title}
                </option>
              ))}
            </select>
            <div className="inline-actions">
              <button
                className="button button-ghost"
                onClick={() => {
                  if (!selectedNoteId) {
                    setStatus('Select a note first.');
                    return;
                  }
                  void generateCardsFromNote(selectedNoteId);
                }}
              >
                Generate Flashcards
              </button>
              <button
                className="button button-ghost"
                onClick={() => {
                  if (!selectedNoteId) {
                    setStatus('Select a note first.');
                    return;
                  }
                  void createQuizFromNote(selectedNoteId);
                }}
              >
                Generate Quiz
              </button>
            </div>
          </div>

          <div className="panel form-stack">
            <h3>Instant Review Card</h3>
            <p className="muted small">
              Pull a due card instantly and review from Command Center.
            </p>
            <button className="button button-ghost" onClick={pickQuickReviewCard}>
              Pick Random Card
            </button>
            {!!quickReviewCard && (
              <div className="panel sub-panel form-stack">
                <p className="label">{quickReviewCard.front}</p>
                <button
                  className="button button-ghost"
                  onClick={() => setShowQuickReviewAnswer((curr) => !curr)}
                >
                  {showQuickReviewAnswer ? 'Hide Answer' : 'Reveal Answer'}
                </button>
                {showQuickReviewAnswer && <p>{quickReviewCard.back}</p>}
                <div className="inline-actions">
                  <button
                    className="button button-ghost"
                    onClick={() => void reviewCard(quickReviewCard.id, 2)}
                  >
                    Hard
                  </button>
                  <button
                    className="button button-ghost"
                    onClick={() => void reviewCard(quickReviewCard.id, 4)}
                  >
                    Good
                  </button>
                  <button
                    className="button button-ghost"
                    onClick={() => void reviewCard(quickReviewCard.id, 5)}
                  >
                    Easy
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </section>
    );
  }

  function renderNotes() {
    return (
      <section className="view-stack">
        <form onSubmit={onCreateNote} className="panel panel-heavy form-stack">
          <h2>Create Smart Note</h2>
          <input
            className="input"
            placeholder="Note title"
            value={noteTitle}
            onChange={(e) => setNoteTitle(e.target.value)}
          />
          <input
            className="input"
            placeholder="Tags (comma separated)"
            value={noteTags}
            onChange={(e) => setNoteTags(e.target.value)}
          />
          <textarea
            className="textarea"
            placeholder="Paste raw notes here"
            value={noteContent}
            onChange={(e) => setNoteContent(e.target.value)}
          />
          <button className="button button-primary" type="submit">
            Save Smart Note
          </button>
        </form>

        <div className="note-grid">
          {filteredNotes.map((note) => (
            <article key={note.id} className="panel">
              <div className="item-head">
                <h3>{note.title}</h3>
                <div className="inline-actions">
                  <button
                    className="button button-ghost"
                    onClick={() => generateCardsFromNote(note.id)}
                  >
                    Make Cards
                  </button>
                  <button
                    className="button button-ghost"
                    onClick={() => createQuizFromNote(note.id)}
                  >
                    Make Quiz
                  </button>
                </div>
              </div>
              <div className="tag-wrap">
                {(note.tags ?? []).map((tag) => (
                  <span key={tag} className="chip">
                    {tag}
                  </span>
                ))}
              </div>
              <p className="muted">{note.summary}</p>
              {!!note.keyTerms?.length && (
                <p className="small">Key terms: {note.keyTerms.join(', ')}</p>
              )}
              {!!note.bulletPoints?.length && (
                <ul className="simple-list">
                  {note.bulletPoints.map((point, idx) => (
                    <li key={`${note.id}-${idx}`}>{point}</li>
                  ))}
                </ul>
              )}
            </article>
          ))}
          {!filteredNotes.length && (
            <p className="muted">No notes match your search.</p>
          )}
        </div>
      </section>
    );
  }

  function renderMnemonics() {
    return (
      <section className="view-stack">
        <form
          onSubmit={onGenerateAcronyms}
          className="panel panel-heavy form-stack"
        >
          <h2>Mnemonic Forge</h2>
          <textarea
            className="textarea"
            placeholder="Enter words or concepts (line/comma separated)"
            value={acronymWords}
            onChange={(e) => setAcronymWords(e.target.value)}
          />
          <button className="button button-primary" type="submit">
            Generate Acronyms
          </button>
        </form>

        <div className="grid-duo">
          {acronyms.map((item) => (
            <article key={item.acronym} className="panel">
              <div className="item-head">
                <h3>{item.acronym}</h3>
                <span className="chip">Score {item.score}</span>
              </div>
              <p className="muted small">
                Readability {item.readabilityScore} | Familiarity{' '}
                {item.familiarityScore}
              </p>
              <p>{item.mnemonic}</p>
              <button
                className="button button-ghost"
                onClick={() => copyMnemonic(item.mnemonic)}
              >
                Copy Mnemonic
              </button>
            </article>
          ))}
          {!acronyms.length && (
            <p className="muted">Generated acronyms will appear here.</p>
          )}
        </div>
      </section>
    );
  }

  function renderFlashcards() {
    return (
      <section className="view-stack">
        <form
          className="panel panel-heavy form-stack"
          onSubmit={onCreateManualCard}
        >
          <h2>Manual Flashcard Builder</h2>
          <label className="toggle-row">
            <input
              type="checkbox"
              checked={dueOnlyFlashcards}
              onChange={(e) => setDueOnlyFlashcards(e.target.checked)}
            />
            Load due-only cards from API
          </label>
          <select
            className="select"
            value={selectedNoteId}
            onChange={(e) => setSelectedNoteId(e.target.value)}
          >
            <option value="">Optional note link</option>
            {notes.map((note) => (
              <option key={note.id} value={note.id}>
                {note.title}
              </option>
            ))}
          </select>
          <input
            className="input"
            placeholder="Front"
            value={manualCardFront}
            onChange={(e) => setManualCardFront(e.target.value)}
          />
          <textarea
            className="textarea"
            placeholder="Back"
            value={manualCardBack}
            onChange={(e) => setManualCardBack(e.target.value)}
          />
          <button className="button button-primary" type="submit">
            Create Flashcard
          </button>
        </form>

        <div className="note-grid">
          {filteredFlashcards.map((card) => (
            <article key={card.id} className="panel">
              <p className="muted small">
                Next: {new Date(card.nextReview).toLocaleString()}
              </p>
              <p className="label">{card.front}</p>
              <p>{card.back}</p>
              <div className="inline-actions">
                <button
                  className="button button-ghost"
                  onClick={() => reviewCard(card.id, 2)}
                >
                  Hard
                </button>
                <button
                  className="button button-ghost"
                  onClick={() => reviewCard(card.id, 4)}
                >
                  Good
                </button>
                <button
                  className="button button-ghost"
                  onClick={() => reviewCard(card.id, 5)}
                >
                  Easy
                </button>
              </div>
            </article>
          ))}
          {!filteredFlashcards.length && (
            <p className="muted">No flashcards match your search.</p>
          )}
        </div>
      </section>
    );
  }

  function renderQuizzes() {
    return (
      <section className="view-stack">
        <div className="panel panel-heavy form-stack">
          <h2>Quiz Arena</h2>
          <select
            className="select"
            value={activeQuizId}
            onChange={(e) => {
              setActiveQuizId(e.target.value);
              setQuizAnswers({});
              setLastQuizScore(null);
              setRandomQuestion(null);
            }}
          >
            <option value="">Select a quiz</option>
            {filteredQuizzes.map((quiz) => (
              <option key={quiz.id} value={quiz.id}>
                {quiz.title}
              </option>
            ))}
          </select>

          <div className="inline-actions">
            <button
              className="button button-ghost"
              onClick={pickRandomQuestion}
            >
              Random Question
            </button>
          </div>

          {!!randomQuestion && (
            <div className="panel sub-panel">
              <p className="label">{randomQuestion.text}</p>
              {!!randomQuestion.options?.length && (
                <p className="muted small">
                  Options: {randomQuestion.options.join(' | ')}
                </p>
              )}
              <button
                className="button button-ghost"
                onClick={() => setShowRandomAnswer((v) => !v)}
              >
                {showRandomAnswer ? 'Hide Answer' : 'Reveal Answer'}
              </button>
              {showRandomAnswer && (
                <p className="small">Answer: {randomQuestion.correctAnswer}</p>
              )}
            </div>
          )}

          {!!activeQuiz && (
            <div className="form-stack">
              {activeQuiz.questions.map((question, index) => (
                <div key={question.id} className="panel sub-panel">
                  <p className="label">
                    {index + 1}. {question.text}
                  </p>
                  {!!question.options?.length ? (
                    <select
                      className="select"
                      value={quizAnswers[question.id] ?? ''}
                      onChange={(e) =>
                        setQuizAnswers((curr) => ({
                          ...curr,
                          [question.id]: e.target.value,
                        }))
                      }
                    >
                      <option value="">Choose an answer</option>
                      {question.options.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      className="input"
                      value={quizAnswers[question.id] ?? ''}
                      onChange={(e) =>
                        setQuizAnswers((curr) => ({
                          ...curr,
                          [question.id]: e.target.value,
                        }))
                      }
                      placeholder="Type your answer"
                    />
                  )}
                </div>
              ))}
              <button className="button button-primary" onClick={submitQuiz}>
                Submit Quiz
              </button>
              {lastQuizScore !== null && (
                <p className="label">Latest Score: {lastQuizScore}%</p>
              )}
            </div>
          )}
        </div>
      </section>
    );
  }

  function renderFocusMode() {
    return (
      <section className="view-stack">
        <div className="panel panel-heavy form-stack">
          <h2>Focus Timer</h2>
          <p className="muted small">
            25 min focus / 5 min break. Auto-saves each completed cycle.
          </p>

          <div className="timer-face">{formatClock(secondsLeft)}</div>

          <input
            className="input"
            value={studyTopic}
            onChange={(e) => setStudyTopic(e.target.value)}
            placeholder="Topic for this session"
          />

          <div className="inline-actions">
            <button
              className="button button-primary"
              onClick={() => setPomodoroOn(true)}
            >
              Start
            </button>
            <button
              className="button button-ghost"
              onClick={() => setPomodoroOn(false)}
            >
              Pause
            </button>
            <button
              className="button button-ghost"
              onClick={() => {
                setPomodoroOn(false);
                setSecondsLeft(25 * 60);
                setMode('focus');
              }}
            >
              Reset
            </button>
          </div>

          <button
            className="button button-primary"
            onClick={() => {
              const saveDuration = mode === 'focus' ? 25 * 60 : 5 * 60;
              if (mode === 'focus') {
                setTodayFocusSeconds((curr) => curr + saveDuration);
              }
              void savePomodoroSession(mode, saveDuration);
            }}
          >
            Save Current Session
          </button>
        </div>
      </section>
    );
  }

  return (
    <div className="ssa-shell">
      <aside className="ssa-sidebar">
        <div className="brand-block">
          <p className="eyebrow">Smart Study Assistant</p>
          <h1>BLACK MODE</h1>
          <p className="muted small">
            Reimagined interface with deeper focus and faster review loops.
          </p>
        </div>

        <div className="panel sub-panel">
          <label htmlFor="global-search" className="small muted">
            Global search
          </label>
          <input
            id="global-search"
            className="input"
            value={globalSearch}
            onChange={(e) => setGlobalSearch(e.target.value)}
            placeholder="Search notes, cards, quizzes"
          />
          <p className="small muted">
            Status: {loading ? 'Working...' : status}
          </p>
        </div>

        <nav className="nav-list">
          {navItems.map((item) => (
            <button
              key={item.id}
              className={`nav-item ${activeView === item.id ? 'active' : ''}`}
              onClick={() => setActiveView(item.id)}
            >
              <span>{item.label}</span>
              <small>{item.hint}</small>
            </button>
          ))}
        </nav>
      </aside>

      <main className="ssa-main">
        {activeView === 'command' && renderCommandCenter()}
        {activeView === 'notes' && renderNotes()}
        {activeView === 'mnemonics' && renderMnemonics()}
        {activeView === 'flashcards' && renderFlashcards()}
        {activeView === 'quizzes' && renderQuizzes()}
        {activeView === 'focus' && renderFocusMode()}
      </main>
    </div>
  );
}
