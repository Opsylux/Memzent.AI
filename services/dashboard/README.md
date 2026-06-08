# Memzent Dashboard

The admin control panel for Memzent. Built with Next.js 16 (App Router), React 19, Tailwind CSS v4, and Shadcn UI.

## Features

- Real-time cache analytics and GPU avoidance metrics
- Tool registry management (CRUD, Qdrant sync status)
- API key management with rotation and TTL
- Billing dashboard with spend limits and budget forecasting
- Workflow registry lifecycle management
- Audit log viewer
- Live playground for testing prompts
- Notification/webhook configuration

## Development

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Build

```bash
npm run build
```

## Architecture

- **App Router only** (`src/app/`) — no `pages/` directory
- **Server Actions** in `src/app/actions.ts` call the Gateway API
- **Tailwind v4** with `@theme inline` in `globals.css`
- **Auth**: Supabase (org membership via `members` table)
- **Path alias**: `@/*` → `./src/*`

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
