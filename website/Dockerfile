# Stage 1: Development / Build
FROM node:24-alpine AS base

WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json package-lock.json* ./
RUN npm install

# Copy source files
COPY . .

# Development stage
FROM base AS dev
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# # Production Build stage
# FROM base AS build
# RUN npm run build

# # Stage 2: Production Serving
# FROM nginx:alpine AS prod
# COPY --from=build /app/dist /usr/share/nginx/html
# EXPOSE 80
# CMD ["nginx", "-g", "daemon off;"]
