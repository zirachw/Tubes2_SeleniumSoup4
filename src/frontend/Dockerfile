FROM node:18-alpine AS base
RUN apk add --no-cache g++ make py3-pip libc6-compat
WORKDIR /app
ARG NODE_ENV=production
ENV NODE_ENV=${NODE_ENV}

ARG NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL}

COPY package*.json ./

FROM base AS deps
RUN npm ci

FROM deps AS builder
COPY . .
RUN npm run build

FROM base AS production
WORKDIR /app
ENV NODE_ENV=production

RUN npm ci --omit=dev

COPY --from=builder /app/.next ./.next
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/public ./public

RUN addgroup -g 1001 -S nodejs \
 && adduser -S nextjs -u 1001 -G nodejs
USER nextjs

EXPOSE 3000
CMD ["npm", "start"]

FROM base AS dev
ENV NODE_ENV=development
RUN npm install
COPY . .
CMD ["npm", "run", "dev"]
