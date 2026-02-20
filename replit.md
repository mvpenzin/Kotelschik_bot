# SNT-Bot

## Overview

SNT-Bot is an automation system for the Russian gardening community (СНТ "Котельщик"). It consists of two main parts:

1. **A Telegram bot backend** (written in Go) that handles member interactions — profile management, plot tracking, payments, weather, train schedules, contacts, entertainment features (jokes, quotes), and admin functions.
2. **A web-based admin dashboard** (React + Express) for managing the bot — viewing status, logs, contacts, pricing, debts, and payment details.

The Express server acts as a reverse proxy, forwarding `/api` requests to the Go backend running on port 8080. The React frontend is a single-page app served by Express.

## User Preferences

Preferred communication style: Simple, everyday language.

## System Architecture

### Frontend (React SPA)
- **Framework**: React with TypeScript, bundled by Vite
- **Routing**: Wouter (lightweight client-side router)
- **UI Components**: shadcn/ui (new-york style) built on Radix UI primitives with Tailwind CSS
- **State/Data Fetching**: TanStack React Query with 5-second polling for live bot status and logs
- **Forms**: React Hook Form with Zod resolvers for validation
- **Styling**: Tailwind CSS with dark theme (HSL CSS variables), Inter font for UI, JetBrains Mono for code/logs
- **Pages**: Dashboard (status overview), Contacts (CRUD management), Logs (filtered log viewer), plus planned pages for Details, Prices, and Debts
- **Path aliases**: `@/` maps to `client/src/`, `@shared/` maps to `shared/`

### Backend (Node.js Express + Go)
- **Express server** (`server/index.ts`): Serves the React frontend and proxies all `/api/*` requests to the Go backend via `http-proxy-middleware`
- **Go backend**: Spawned as a child process from Express. Handles all API logic and Telegram bot functionality. Runs on port 8080.
- **In-memory storage** (`server/storage.ts`): A simple `MemStorage` class exists for Express-side user management, but the primary data lives in PostgreSQL accessed by the Go backend.

### Shared Code (`shared/`)
- **Schema** (`shared/schema.ts`): Drizzle ORM table definitions shared between frontend validation and backend. Tables include:
  - `snt_users` — Telegram users (telegram_id, username, first/last name)
  - `snt_contacts` — Community contacts (type, value, comment, priority)
  - `bot_logs` — Bot activity logs (level, message, details)
- **Routes** (`shared/routes.ts`): API contract definitions with Zod schemas for request/response validation. Defines endpoints for status, logs, and contacts CRUD.

### Database
- **PostgreSQL** via Drizzle ORM
- **Schema management**: `drizzle-kit push` for migrations (no migration files, direct push)
- **Connection**: `DATABASE_URL` environment variable required
- **Session store**: `connect-pg-simple` is included as a dependency

### Build System
- **Development**: `tsx` runs the Express server directly; Vite dev server with HMR handles the frontend
- **Production**: Vite builds the React app to `dist/public`; esbuild bundles the Express server to `dist/index.cjs`
- **The build script** bundles common server dependencies (express, drizzle-orm, pg, etc.) to reduce cold start times

### Telegram Bot Features (Go backend)
Based on the requirements document, the bot provides:
- User registration on any interaction
- Profile management (name, phone, birthday, plot numbers)
- Information services (weather, train schedules, contacts, pricing)
- Payment stubs (Sberbank, QR codes)
- Entertainment (holidays, quotes, jokes from anekdot.ru, bash.im)
- Auto-closing keyboards after 60 seconds of inactivity
- Admin keyboard for management

## External Dependencies

- **PostgreSQL**: Primary database, required via `DATABASE_URL` environment variable
- **Telegram Bot API**: Core functionality of the Go backend (bot token needed)
- **Go runtime**: Required to compile and run the bot backend
- **External content sources** (used by the bot):
  - Weather API (for weather forecasts)
  - Yandex train schedules (`rasp.yandex.ru`)
  - anekdot.ru (jokes)
  - bash.im (quotes)
- **npm packages of note**:
  - `http-proxy-middleware` — proxies API calls from Express to Go
  - `drizzle-orm` + `drizzle-kit` — ORM and migration tooling
  - `@tanstack/react-query` — data fetching with polling
  - `react-hook-form` + `zod` — form handling and validation
  - Full shadcn/ui component library (Radix-based)