FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod ./
COPY backend/ ./
RUN go mod tidy && CGO_ENABLED=0 go build -o /music-backend .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=backend-builder /music-backend /music-backend
COPY --from=frontend-builder /frontend/dist /app/frontend/dist
EXPOSE 3000
CMD ["/music-backend"]
