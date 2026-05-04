# syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm install --no-audit --no-fund

FROM node:22-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22-alpine AS runtime
RUN addgroup -S workend && adduser -S -G workend -u 10001 workend
USER workend
WORKDIR /app
COPY --from=build --chown=workend:workend /app/build ./build
COPY --from=build --chown=workend:workend /app/node_modules ./node_modules
COPY --from=build --chown=workend:workend /app/package.json ./package.json
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000
CMD ["node", "build/index.js"]
