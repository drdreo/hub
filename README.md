# Hub

**One monorepo to rule them all.**

Hub is the single home for my side projects — web games, web servers, and CLIs — built around **one generic game server** that any client (web, mobile, CLI) can plug into, regardless of stack.

## Goal

- **Stop reinventing infrastructure.** Connection management, rooms, reconnection, message routing, and lobby state are solved *once* in `apps/gameserver` (Go, WebSocket, plugin-based) and reused by every game.
- **Games are plugins.** A new game implements a small `Game` interface (`HandleMessage`, `InitializeRoom`, `OnClientJoin/Leave`) and registers itself. The server handles the rest.
- **Clients are interchangeable.** Any frontend or CLI that speaks the wire protocol (`{ type, data }` envelopes over WebSocket) can join a room.
- **Shared tooling.** Nx orchestrates builds, lint, and tests across Go and TypeScript. Each app stays deployable on its own (Cloudflare for the web clients, Docker/Railway for the APIs).

## Layout

```
apps/
├── gameserver/      # Go WebSocket server — the generic core (rooms, sessions, registry)
│   └── games/       # Plugin games: dicegame, owe_drahn, tictactoe, tell_it
├── owe-drahn/       # React + Vite web client (deployed to Cloudflare Pages)
├── tell-it-api/     # NestJS API (legacy, pre-gameserver)
└── demo-cli/        # Go CLI client to automate demos

packages/
└── tell-it/         # Shared TS libs (api / shared / web)
```

See [`apps/gameserver/README.md`](./apps/gameserver/README.md) for the wire protocol, room lifecycle, and reconnection flow.


### Yet to be migrated projects

- **Poker** → A texas hold'em variant https://github.com/drdreo/poker
- **KCDice** → A farkle KCD2 inpsired dice game https://github.com/drdreo/dicegame (Angular frontend migration blocked by TS reference vs alias support)

## Common commands

```sh
nx affected -t build           # build everything touched by current changes
nx affected -t test            # test everything touched
nx affected -t lint
nx format:check                # or: nx format:write

nx test <project>              # single project, e.g. nx test owe-drahn
nx <target> <project>          # generic form

# Go-specific (gameserver / demo-cli)
cd apps/gameserver && go test ./...
```

Targets are either inferred by Nx or defined per-project in `project.json` / `package.json`.

## Deploying

```sh
nx deploy <project>
```

- **Web clients** → Cloudflare (via the local `@hub/plugin:cloudflare` executor and `wrangler`).
- **APIs** → Docker images pushed to GHCR, deployed to Railway (via `@hub/plugin:deploy-docker`).

## TypeScript project references

Nx keeps `tsconfig.json` references in sync with the import graph automatically on `build` / `typecheck`. To force a sync or verify in CI:

```sh
npx nx sync          # write
npx nx sync:check    # verify (CI)
```
