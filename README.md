# Hub

<a alt="Nx logo" href="https://nx.dev" target="_blank" rel="noreferrer"><img src="https://raw.githubusercontent.com/nrwl/nx/master/images/nx-logo.png" width="45"></a>

✨ Your new, shiny [Nx workspace](https://nx.dev) is almost ready ✨.

[Learn more about this workspace setup and its capabilities](https://nx.dev/nx-api/js?utm_source=nx_project&utm_medium=readme&utm_campaign=nx_projects) or run `npx nx graph` to visually explore what was created. Now, let's get you up to speed!

### Folder Structure

```
libs/
├── tell-it/
│   ├── web/                # Angular-specific libraries
│   │   ├── feature/        # Angular feature libraries
│   │   ├── ui/             # Angular UI components
│   │   └── data-access/    # Angular services
│   ├── api/                # NestJS-specific libraries
│   │   ├── feature/        # NestJS feature modules
│   │   └── data-access/    # NestJS services, repositories
│   └── shared/             # Shared between web and api
│       ├── models/         # Shared interfaces/types
│       └── util/           # Shared utilities
├── shared/                 # Cross-domain shared code
```

To run any task with Nx use:

```sh
npx nx <target> <project-name>
```

These targets are either [inferred automatically](https://nx.dev/concepts/inferred-tasks?utm_source=nx_project&utm_medium=readme&utm_campaign=nx_projects) or defined in the `project.json` or `package.json` files.

[More about running tasks in the docs &raquo;](https://nx.dev/features/run-tasks?utm_source=nx_project&utm_medium=readme&utm_campaign=nx_projects)

## Deploying

To deploy an application:

```
nx deploy
```

For APIs, dockerfiles are used to generate docker images.
Make sure to correctly setup the registry to allow the repo to push images.
Manage package settings > Choose repo > Role: Admin

[Learn more about Nx release &raquo;](hhttps://nx.dev/features/manage-releases?utm_source=nx_project&utm_medium=readme&utm_campaign=nx_projects)

## Keep TypeScript project references up to date

Nx automatically updates TypeScript [project references](https://www.typescriptlang.org/docs/handbook/project-references.html) in `tsconfig.json` files to ensure they remain accurate based on your project dependencies (`import` or `require` statements). This sync is automatically done when running tasks such as `build` or `typecheck`, which require updated references to function correctly.

To manually trigger the process to sync the project graph dependencies information to the TypeScript project references, run the following command:

```sh
npx nx sync
```

You can enforce that the TypeScript project references are always in the correct state when running in CI by adding a step to your CI job configuration that runs the following command:

```sh
npx nx sync:check
```

[Learn more about nx sync](https://nx.dev/reference/nx-commands#sync)
