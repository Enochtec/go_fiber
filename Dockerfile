FROM node:22-bookworm AS frontend

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build && rm -rf web && cp -r build web

FROM golang:1.22-bookworm AS backend

WORKDIR /app
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend ./backend
COPY --from=frontend /app/web ./web
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -o /pos ./cmd

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=backend /pos ./pos
COPY --from=backend /app/web ./web

EXPOSE 10000
CMD ["./pos"]