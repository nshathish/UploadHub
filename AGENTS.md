# Repository Guidelines

## Project Structure & Module Organization

This repository has two independently runnable modules:

- `frontend/` is a React 19, TypeScript, and Vite application. UI entry points
  live in `src/` (`main.tsx`, `App.tsx`); component styles are colocated or in
  `src/index.css`; static assets belong in `src/assets/` or `public/`.
- `backend/` is the Go module (`github.com/nshathish/uploadhub/backend`). Add
  server packages beneath this directory; keep package tests next to the code
  they exercise as `*_test.go`.

Avoid committing generated output such as `frontend/dist/`, dependency folders,
or local environment files.

## Build, Test, and Development Commands

Run frontend commands from `frontend/`:

- `pnpm dev` starts Vite with hot module replacement.
- `pnpm build` type-checks with TypeScript and creates the production bundle.
- `pnpm lint` runs ESLint over TypeScript and TSX files.
- `pnpm preview` serves the production build locally.

Run backend commands from `backend/`:

- `go test ./...` runs all Go tests (and is the expected test command as the
  backend grows).
- `go build ./...` verifies every Go package compiles.
- `gofmt -w .` formats changed Go source files; prefer targeting files or
  packages when the codebase expands.

## Coding Style & Naming Conventions

Use the existing TypeScript style: two-space indentation, single quotes,
semicolon-free statements, trailing commas, and functional React components.
Name components in PascalCase (`UploadPanel.tsx`), hooks with `use` prefixes,
and ordinary variables/functions in camelCase. Keep imports explicit and use
the ESLint configuration in `frontend/eslint.config.js` as the source of truth.

For Go, follow standard `gofmt` formatting. Use concise lowercase package
names, exported PascalCase identifiers, and descriptive file names such as
`upload_handler.go`.

## Testing Guidelines

No test framework is configured in the frontend yet. Add one with the related
feature and document its command in `frontend/package.json`. For Go, write
table-driven tests where useful, name files `*_test.go`, and name tests
`Test<Behavior>` (for example, `TestUploadRejectsOversizeFile`). Run lint and
the relevant build/test command before opening a pull request.

## Commit & Pull Request Guidelines

The current history contains only `Initial commit`, so no established commit
format exists. Use short, imperative summaries, preferably scoped, such as
`frontend: add upload dropzone` or `backend: validate file size`. Keep commits
focused. Pull requests should explain the change, note validation performed,
link the relevant issue, and include screenshots for visible UI changes.
