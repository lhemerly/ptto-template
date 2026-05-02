# Contributing to the ptto Go Template

This repository serves as the baseline for all `ptto new` commands. Because it is a template, **simplicity and auditability are our highest priorities.**

Before opening a PR, understand our strict architectural boundaries:

## 🛑 What we will reject automatically:
1. **Adding CGO dependencies**: We use `modernc.org/sqlite` specifically so this app can be easily cross-compiled (`GOOS=linux go build`) without needing a C-compiler toolchain. Do not add libraries that wrap C code.
2. **Adding SPA Frameworks**: No React, Vue, Svelte, or Next.js. We return HTML over the wire. Period.
3. **Adding Third-Party Auth**: Do not add OAuth, Google Auth, or Magic Link providers. 
4. **Adding an ORM**: We use standard SQL queries. No heavy, reflection-based ORMs.

## Code Style
Keep the Go code idiomatic and boring. Avoid deep interface abstractions unless absolutely necessary. Handlers should parse the request, query the database, and render the `templ` component.
