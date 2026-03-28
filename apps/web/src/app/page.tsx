'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';

type View =
  | 'dashboard'
  | 'notes'
  | 'acronyms'
  | 'flashcards'
  | 'quizzes'
  | 'study';

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

export default function Home() {
  const [activeView, setActiveView] = useState<View>('dashboard');
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState('Ready');

  const [userId, setUserId] = useState(DEFAULT_USER_ID);

  const [notes, setNotes] = useState<Note[]>([]);
  const [flashcards, setFlashcards] = useState<Flashcard[]>([]);
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);

  const [noteTitle, setNoteTitle] = useState('');
  const [noteTags, setNoteTags] = useState('');
  const [noteContent, setNoteContent] = useState('');

  const [acronymWords, setAcronymWords] = useState('');
  const [acronyms, setAcronyms] = useState<AcronymSuggestion[]>([]);

  const [manualCardFront, setManualCardFront] = useState('');
  const [manualCardBack, setManualCardBack] = useState('');
  const [selectedNoteId, setSelectedNoteId] = useState('');

  const [activeQuizId, setActiveQuizId] = useState('');
  const [quizAnswers, setQuizAnswers] = useState<Record<string, string>>({});
  const [lastQuizScore, setLastQuizScore] = useState<number | null>(null);

  const [pomodoroOn, setPomodoroOn] = useState(false);
  const [secondsLeft, setSecondsLeft] = useState(25 * 60);
  const [mode, setMode] = useState<'focus' | 'break'>('focus');
  const [studyTopic, setStudyTopic] = useState('General Study');

  const activeQuiz = useMemo(
    () => quizzes.find((quiz) => quiz.id === activeQuizId),
    [quizzes, activeQuizId],
  );

  const navItems: { id: View; label: string }[] = [
    { id: 'dashboard', label: 'Progress' },
    { id: 'notes', label: 'Notes' },
    { id: 'acronyms', label: 'Acronyms' },
    { id: 'flashcards', label: 'Flashcards' },
    { id: 'quizzes', label: 'Quizzes' },
    { id: 'study', label: 'Study Mode' },
  ];

  async function loadAll() {
    setLoading(true);
    try {
      const [noteData, flashcardData, quizData, dashboardData] =
        await Promise.all([
          gql<{ notes: Note[] }>(
            `query($userId: String!) { notes(userId: $userId) { id title content summary keyTerms bulletPoints examReadyText tags } }`,
            { userId },
          ),
          gql<{ flashcards: Flashcard[] }>(
            `query($userId: String!) { flashcards(userId: $userId, dueOnly: false) { id noteId front back nextReview interval easeFactor repetition } }`,
            { userId },
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
  }

  useEffect(() => {
    loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]);

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
      const next = prev === 'focus' ? 'break' : 'focus';
      setSecondsLeft(next === 'focus' ? 25 * 60 : 5 * 60);
      return next;
    });
  }, [pomodoroOn, secondsLeft]);

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

  async function onCreateManualCard(e: FormEvent) {
    e.preventDefault();
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

  async function savePomodoroSession() {
    const minutes = mode === 'focus' ? 25 : 5;
    setLoading(true);
    try {
      await gql(
        `mutation($userId: String!, $duration: Int!, $type: String!, $topic: String) {
          saveStudySession(userId: $userId, duration: $duration, type: $type, topic: $topic) { id }
        }`,
        {
          userId,
          duration: minutes * 60,
          type: mode === 'focus' ? 'pomodoro' : 'break',
          topic: studyTopic,
        },
      );
      setStatus('Study session saved.');
      await loadAll();
    } catch (error) {
      setStatus(
        error instanceof Error ? error.message : 'Could not save study session',
      );
    } finally {
      setLoading(false);
    }
  }

  function renderDashboard() {
    const totalHours = ((dashboard?.totalStudySeconds ?? 0) / 3600).toFixed(1);

    return (
      <section className="panel p-5 space-y-4">
        <h2 className="text-2xl font-bold">Progress Dashboard</h2>
        <div className="grid md:grid-cols-3 gap-3">
          <div className="panel p-3">
            <p className="muted text-xs uppercase">Study Time</p>
            <p className="text-2xl font-semibold">{totalHours}h</p>
          </div>
          <div className="panel p-3">
            <p className="muted text-xs uppercase">Quiz Attempts</p>
            <p className="text-2xl font-semibold">
              {dashboard?.quizAttempts ?? 0}
            </p>
          </div>
          <div className="panel p-3">
            <p className="muted text-xs uppercase">Average Quiz Score</p>
            <p className="text-2xl font-semibold">
              {dashboard?.averageQuizScore ?? 0}%
            </p>
          </div>
        </div>

        <div className="grid md:grid-cols-2 gap-3">
          <div className="panel p-3">
            <p className="font-semibold mb-2">Topics Covered</p>
            <div className="flex flex-wrap gap-2">
              {(dashboard?.topicsCovered ?? []).map((topic) => (
                <span key={topic} className="chip">
                  {topic}
                </span>
              ))}
              {!dashboard?.topicsCovered?.length && (
                <p className="muted text-sm">No topics yet.</p>
              )}
            </div>
          </div>
          <div className="panel p-3">
            <p className="font-semibold mb-2">Weak Areas</p>
            <div className="space-y-2">
              {(dashboard?.weakAreas ?? []).map((topic) => (
                <div key={topic} className="flex items-center justify-between">
                  <span>{topic}</span>
                  <span className="text-xs px-2 py-1 rounded bg-red-100 text-red-700">
                    Review
                  </span>
                </div>
              ))}
              {!dashboard?.weakAreas?.length && (
                <p className="muted text-sm">No weak areas recorded.</p>
              )}
            </div>
          </div>
        </div>
      </section>
    );
  }

  function renderNotes() {
    return (
      <section className="space-y-4">
        <form onSubmit={onCreateNote} className="panel p-5 space-y-3">
          <h2 className="text-2xl font-bold">Smart Notes Generator</h2>
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

        <div className="space-y-3">
          {notes.map((note) => (
            <article key={note.id} className="panel p-4 space-y-2">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h3 className="text-lg font-semibold">{note.title}</h3>
                  <div className="flex gap-2 flex-wrap mt-1">
                    {(note.tags ?? []).map((tag) => (
                      <span key={tag} className="chip">
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    className="button button-secondary"
                    onClick={() => generateCardsFromNote(note.id)}
                  >
                    Flashcards
                  </button>
                  <button
                    className="button button-secondary"
                    onClick={() => createQuizFromNote(note.id)}
                  >
                    Quiz
                  </button>
                </div>
              </div>
              <p className="muted">{note.summary}</p>
              {!!note.keyTerms?.length && (
                <p>
                  <span className="font-semibold">Key terms:</span>{' '}
                  {note.keyTerms.join(', ')}
                </p>
              )}
              {!!note.bulletPoints?.length && (
                <ul className="list-disc pl-5 space-y-1">
                  {note.bulletPoints.map((point, idx) => (
                    <li key={`${note.id}-${idx}`}>{point}</li>
                  ))}
                </ul>
              )}
            </article>
          ))}
          {!notes.length && (
            <p className="muted">
              Create your first note to unlock auto-generated study assets.
            </p>
          )}
        </div>
      </section>
    );
  }

  function renderAcronyms() {
    return (
      <section className="space-y-4">
        <form onSubmit={onGenerateAcronyms} className="panel p-5 space-y-3">
          <h2 className="text-2xl font-bold">Acronym & Mnemonic Generator</h2>
          <textarea
            className="textarea"
            placeholder="Enter words, bullets, or concepts (one per line)"
            value={acronymWords}
            onChange={(e) => setAcronymWords(e.target.value)}
          />
          <button className="button button-primary" type="submit">
            Generate Acronyms
          </button>
        </form>

        <div className="grid lg:grid-cols-2 gap-3">
          {acronyms.map((item) => (
            <article key={item.acronym} className="panel p-4 space-y-2">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold tracking-wide">
                  {item.acronym}
                </h3>
                <span className="chip">Score: {item.score}</span>
              </div>
              <p className="muted text-sm">
                Readability {item.readabilityScore} · Familiarity{' '}
                {item.familiarityScore}
              </p>
              <p>{item.mnemonic}</p>
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
      <section className="space-y-4">
        <form className="panel p-5 space-y-3" onSubmit={onCreateManualCard}>
          <h2 className="text-2xl font-bold">Flashcards</h2>
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
            placeholder="Front (question/acronym)"
            value={manualCardFront}
            onChange={(e) => setManualCardFront(e.target.value)}
          />
          <textarea
            className="textarea"
            placeholder="Back (answer/explanation)"
            value={manualCardBack}
            onChange={(e) => setManualCardBack(e.target.value)}
          />
          <button className="button button-primary" type="submit">
            Create Flashcard
          </button>
        </form>

        <div className="grid md:grid-cols-2 gap-3">
          {flashcards.map((card) => (
            <article key={card.id} className="panel p-4 space-y-2">
              <p className="text-sm muted">
                Next review: {new Date(card.nextReview).toLocaleString()}
              </p>
              <p className="font-semibold">{card.front}</p>
              <p>{card.back}</p>
              <div className="flex gap-2">
                <button
                  className="button button-secondary"
                  onClick={() => reviewCard(card.id, 2)}
                >
                  Hard
                </button>
                <button
                  className="button button-secondary"
                  onClick={() => reviewCard(card.id, 4)}
                >
                  Good
                </button>
                <button
                  className="button button-secondary"
                  onClick={() => reviewCard(card.id, 5)}
                >
                  Easy
                </button>
              </div>
            </article>
          ))}
          {!flashcards.length && <p className="muted">No flashcards yet.</p>}
        </div>
      </section>
    );
  }

  function renderQuizzes() {
    return (
      <section className="space-y-4">
        <div className="panel p-5 space-y-3">
          <h2 className="text-2xl font-bold">Quiz Generator</h2>
          <p className="muted text-sm">
            Generate a quiz from any note in the Notes section, then select it
            below to answer.
          </p>
          <select
            className="select"
            value={activeQuizId}
            onChange={(e) => {
              setActiveQuizId(e.target.value);
              setQuizAnswers({});
              setLastQuizScore(null);
            }}
          >
            <option value="">Select a quiz</option>
            {quizzes.map((quiz) => (
              <option key={quiz.id} value={quiz.id}>
                {quiz.title}
              </option>
            ))}
          </select>

          {!!activeQuiz && (
            <div className="space-y-4">
              {activeQuiz.questions.map((question, index) => (
                <div key={question.id} className="panel p-3 space-y-2">
                  <p className="font-medium">
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
                <p className="text-lg font-semibold">
                  Latest Score: {lastQuizScore}%
                </p>
              )}
            </div>
          )}
        </div>
      </section>
    );
  }

  function renderStudyMode() {
    const minutes = Math.floor(secondsLeft / 60);
    const seconds = secondsLeft % 60;

    return (
      <section className="panel p-5 space-y-4">
        <h2 className="text-2xl font-bold">Study Mode</h2>
        <p className="muted">
          Pomodoro timer: 25 minutes focus, 5 minutes break.
        </p>

        <div className="panel p-4 text-center space-y-2">
          <p className="text-xs uppercase muted">
            {mode === 'focus' ? 'Focus Session' : 'Break Session'}
          </p>
          <p className="text-6xl font-bold tabular-nums">
            {String(minutes).padStart(2, '0')}:
            {String(seconds).padStart(2, '0')}
          </p>
        </div>

        <input
          className="input"
          value={studyTopic}
          onChange={(e) => setStudyTopic(e.target.value)}
          placeholder="Study topic"
        />

        <div className="grid md:grid-cols-3 gap-3">
          <button
            className="button button-primary"
            onClick={() => setPomodoroOn(true)}
          >
            Start
          </button>
          <button
            className="button button-secondary"
            onClick={() => setPomodoroOn(false)}
          >
            Pause
          </button>
          <button
            className="button button-secondary"
            onClick={() => {
              setPomodoroOn(false);
              setSecondsLeft(25 * 60);
              setMode('focus');
            }}
          >
            Reset
          </button>
        </div>

        <button className="button button-primary" onClick={savePomodoroSession}>
          Save Session
        </button>
      </section>
    );
  }

  return (
    <div className="study-shell">
      <aside className="study-sidebar">
        <div className="space-y-3">
          <p className="text-xs uppercase muted tracking-[0.2em]">
            Smart Study Assistant
          </p>
          <h1 className="text-2xl font-bold">Focus Console</h1>
          <p className="text-sm muted">
            Minimal and distraction-light workspace for active recall.
          </p>
        </div>

        <div className="mt-4 space-y-2">
          <label className="text-xs uppercase muted">User Id</label>
          <input
            className="input"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
          />
        </div>

        <nav className="mt-6 space-y-2">
          {navItems.map((item) => (
            <button
              key={item.id}
              className={`button text-left ${activeView === item.id ? 'button-primary' : 'button-secondary'}`}
              onClick={() => setActiveView(item.id)}
            >
              {item.label}
            </button>
          ))}
        </nav>

        <div className="mt-6 text-sm muted">
          <p>Status: {loading ? 'Working...' : status}</p>
        </div>
      </aside>

      <main className="study-main">
        {activeView === 'dashboard' && renderDashboard()}
        {activeView === 'notes' && renderNotes()}
        {activeView === 'acronyms' && renderAcronyms()}
        {activeView === 'flashcards' && renderFlashcards()}
        {activeView === 'quizzes' && renderQuizzes()}
        {activeView === 'study' && renderStudyMode()}
      </main>
    </div>
  );
}
