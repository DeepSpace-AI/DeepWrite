# Branching Strategy

## Branches

- `main`: production branch. Only accepts changes through pull requests.
- `develop`: integration branch for day-to-day development.
- `feature/*`: feature branches created from `develop`.
- `hotfix/*`: urgent fixes created from `main`.

## Workflow

1. Create feature branches from `develop`.
2. Merge completed features back into `develop` via PR.
3. Promote stable changes from `develop` to `main` through PRs.
4. Use `hotfix/*` only for urgent production issues, then merge back to both `main` and `develop`.

## Branch Protection Rules

- Protect `main` and `develop`.
- Require pull requests for merges.
- Require at least one review before merge.
- Require CI checks to pass before merge.
- Disallow direct pushes to protected branches.
- Keep branch history linear when possible.
