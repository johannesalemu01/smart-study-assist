# Smart Study Assistant (Monorepo)

## Tech
- Frontend: Next.js (App Router)
- Backend: Go + Gin + GraphQL
- Storage: MongoDB (to be added in later milestones)

## Setup
1. Install dependencies:
   - `npm install`
2. Create env vars:
   - Add `PORT` for the API if you want (default `8080`)
   - Later milestones will introduce `MONGODB_URI` and `OPENAI_API_KEY`

## Run (dev)
- `npm run dev`

This runs:
- `apps/web` on the Next.js dev server
- `apps/api` on `/graphql` and `/healthz`

