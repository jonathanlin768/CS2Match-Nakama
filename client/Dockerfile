# ==============================================
# CS2Match Frontend — 多阶段 Docker 构建
# ==============================================
# Stage 1: 编译 React 应用
FROM node:22-alpine AS builder

WORKDIR /app

# 利用 Docker 缓存: 先安装依赖
COPY package.json package-lock.json ./
RUN npm ci

# 复制源码并构建
COPY . .
# Docker 模式: 前端 JS 在浏览器运行，通过宿主机端口映射连接 Nakama
ARG VITE_NAKAMA_HOST=localhost
ARG VITE_NAKAMA_PORT=7350
ARG VITE_NAKAMA_SERVER_KEY=defaultkey
ARG VITE_NAKAMA_USE_SSL=false
RUN npm run build

# Stage 2: Nginx 托管静态资源
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制自定义 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
