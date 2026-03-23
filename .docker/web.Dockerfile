FROM node:22-alpine AS base

WORKDIR /app

FROM base AS prod-deps
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/
COPY apps/admin/package.json apps/admin/
RUN corepack enable pnpm && pnpm install --frozen-lockfile --prod

FROM base AS builder
COPY --from=prod-deps /app/node_modules ./node_modules
COPY . .
RUN pnpm --filter @deepwrite/web build

FROM base AS runner
ENV NODE_ENV=production
COPY --from=builder /app/apps/web/dist ./dist
COPY --from=builder /app/apps/web/package.json ./package.json

EXPOSE 5173

CMD ["npx", "serve", "-s", "dist", "-l", "5173"]
