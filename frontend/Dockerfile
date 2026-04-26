FROM postgres:18-alpine AS builder

WORKDIR /app

RUN apk add --no-cache nodejs npm

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM postgres:18-alpine

WORKDIR /app

RUN apk add --no-cache nodejs npm

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

CMD ["node", "server.js"]
