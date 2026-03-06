# 构建前端
FROM node:16-alpine as frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# 构建后端
FROM golang:1.18-alpine as backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o backup-system .

# 最终镜像
FROM alpine:latest
WORKDIR /app

# 复制前端构建结果
COPY --from=frontend-builder /app/frontend/dist ./frontend

# 复制后端可执行文件
COPY --from=backend-builder /app/backend/backup-system ./

# 复制配置文件目录
COPY backend/config ./config

# 创建备份目录
RUN mkdir -p backup

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./backup-system"]