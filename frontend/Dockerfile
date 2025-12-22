FROM node:lts-alpine as build-stage
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
# 跳过类型检查，直接构建生产静态资源（vite 本身不阻塞 TS 错误）
RUN npm run build-only

FROM nginx:stable-alpine as production-stage
COPY --from=build-stage /app/dist /usr/share/nginx/html

RUN mkdir -p /etc/nginx/ssl

EXPOSE 80
EXPOSE 443
CMD ["nginx", "-g", "daemon off;"]