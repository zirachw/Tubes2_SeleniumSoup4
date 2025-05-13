FROM node:23-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1

FROM base AS deps
COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* .npmrc* ./
RUN \
  if [ -f yarn.lock ];       then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else echo "lockfile not found" && exit 1; \
  fi

FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN \
  if [ -f yarn.lock ];       then yarn add -D tailwindcss postcss autoprefixer; \
  elif [ -f package-lock.json ]; then npm install -D tailwindcss postcss autoprefixer; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm add -D tailwindcss postcss autoprefixer; \
  fi
RUN mkdir -p src/components

RUN \
  if [ -f yarn.lock ];       then yarn build; \
  elif [ -f package-lock.json ]; then npm run build; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm run build; \
  fi

FROM node:23-alpine AS production
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=builder /app/package.json       ./package.json
COPY --from=builder /app/yarn.lock*         ./yarn.lock
COPY --from=builder /app/package-lock.json* ./package-lock.json
COPY --from=builder /app/pnpm-lock.yaml*    ./pnpm-lock.yaml

RUN \
  if [ -f yarn.lock ];        then yarn --frozen-lockfile --production; \
  elif [ -f package-lock.json ]; then npm ci --only=production;         \
  elif [ -f pnpm-lock.yaml ];    then corepack enable pnpm && pnpm i --frozen-lockfile --prod; \
  else echo "no lockfile!" && exit 1; \
  fi

COPY --from=builder /app/.next    ./.next
COPY --from=builder /app/public   ./public

RUN addgroup --system --gid 1001 nodejs \
 && adduser  --system --uid 1001 nextjs \
 && chown -R nextjs:nodejs /app

USER nextjs

EXPOSE 3000
CMD ["npm", "run", "start"]


FROM base AS dev
WORKDIR /app

COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* .npmrc* ./
RUN \
  if [ -f yarn.lock ];       then yarn; \
  elif [ -f package-lock.json ]; then npm install; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm install; \
  else echo " lockfile not found" && exit 1; \
  fi

RUN \
  if [ -f yarn.lock ];       then yarn add -D tailwindcss postcss autoprefixer; \
  elif [ -f package-lock.json ]; then npm install -D tailwindcss postcss autoprefixer; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm add -D tailwindcss postcss autoprefixer; \
  fi

RUN mkdir -p src/components

ENV NODE_ENV=development
EXPOSE 3000
CMD ["npm", "run", "dev"]
