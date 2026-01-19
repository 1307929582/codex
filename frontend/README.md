# Codex Gateway Frontend

Next.js dashboard for Codex Gateway API.

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **UI**: Tailwind CSS + shadcn/ui components
- **State Management**:
  - TanStack Query (server state)
  - Zustand (auth state)
- **Forms**: React Hook Form + Zod
- **HTTP Client**: Axios
- **Charts**: Recharts

## Getting Started

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.local .env.local
# Edit .env.local and set NEXT_PUBLIC_API_URL
```

### 3. Run Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000)

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
│   ├── (auth)/            # Auth pages (login, register)
│   ├── (dashboard)/       # Dashboard pages
│   └── layout.tsx         # Root layout
├── components/
│   └── ui/                # Reusable UI components
├── lib/
│   ├── api/               # API client
│   ├── stores/            # Zustand stores
│   └── utils.ts           # Utility functions
└── types/
    └── api.ts             # TypeScript types
```

## Features

- ✅ User authentication (JWT)
- ✅ Dashboard with stats
- ✅ API key management
- ✅ Usage logs with pagination
- ✅ Account balance and transactions
- ✅ Responsive design

## Build for Production

```bash
npm run build
npm start
```

## Environment Variables

- `NEXT_PUBLIC_API_URL`: Backend API URL (default: http://localhost:8080)
