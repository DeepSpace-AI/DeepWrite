FROM node:22-alpine AS base

WORKDIR /app

FROM base AS deps
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml turbo.json ./
COPY apps/web/package.json apps/web/
COPY apps/admin/package.json apps/admin/
RUN corepack enable pnpm && pnpm install --frozen-lockfile

FROM base AS builder
RUN corepack enable pnpm
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/package.json ./package.json
COPY --from=deps /app/turbo.json ./turbo.json
COPY --from=deps /app/pnpm-workspace.yaml ./pnpm-workspace.yaml
COPY . .
RUN pnpm --filter @deepwrite/admin build

FROM base AS runner
ENV NODE_ENV=production
COPY --from=builder /app/apps/admin/dist ./dist
COPY --from=builder /app/apps/admin/package.json ./package.json

EXPOSE 5174

CMD ["npx", "serve", "-s", "dist", "-l", "5174"]
