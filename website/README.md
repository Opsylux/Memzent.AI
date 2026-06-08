# Memzent Website

Public marketing site for [memzent.ai](https://memzent.ai). Built with Vite + React 19 + TypeScript + Tailwind CSS.

## Features

- Landing page with product overview and comparison table
- Blog at `/blog` (markdown-based, build-time rendered)
- SEO: sitemap.xml, robots.txt, JSON-LD structured data, OG meta tags
- Responsive mobile-first design

## Development

```bash
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173).

## Build

```bash
npm run build
```

Output goes to `dist/`. Served by nginx in production with SPA fallback routing.

## Blog

Blog posts live in `src/content/blog/*.md` with YAML frontmatter:

```yaml
---
title: "Post Title"
date: "2026-06-01"
excerpt: "Short description"
author: "Author Name"
category: "engineering"
---
```

Posts are loaded at build time via `import.meta.glob`.
