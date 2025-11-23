# Agent Guidelines for Hub Repository

This is my hub NX monorepo containing common tooling for my applications.
From web games, web servers to CLIs. Prefered languages: Go, TypeScript, Angular / React

## Build/Test Commands

-   **Build all**: `nx affected -t build`
-   **Test all**: `nx affected -t test`
-   **Lint all**: `nx affected -t lint`
-   **Format check**: `nx format:check`
-   **Format write**: `nx format:write`
-   **Single project test**: `nx test <project-name>` (e.g., `nx test owe-drahn`)
-   **Single Go test**: `cd apps/gameserver && go test ./path/to/package -run TestName`
-   **Go test file**: `cd apps/gameserver && go test ./path/to/package -v`

## Code Style

-   **TypeScript**: Strict mode enabled, no implicit any, decorators enabled for NestJS
-   **Imports**: Use path aliases from tsconfig, absolute imports preferred over relative beyond parent directory
-   **Formatting**: Prettier with 104 print width, 4 spaces, no trailing commas, arrow parens avoided
-   **Naming**: camelCase for TS/JS, PascalCase for React components/Go types, snake_case for Go files
-   **Error handling**: Use class-validator for DTOs, proper error responses with success/error flags
-   **Testing**: Vitest for TS, Go native testing, descriptive test names, mock external dependencies
-   **React**: Functional components with hooks, SCSS modules for styling, Redux Toolkit for state
-   **Go**: Interfaces in internal/interfaces, table-driven tests, proper error wrapping
-   **Monorepo**: Nx workspace, apps/ for deployables, packages/ for shared libs, enforce module boundaries
