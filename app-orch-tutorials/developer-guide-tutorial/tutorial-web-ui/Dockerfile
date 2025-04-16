# Stage 1: Build the application
FROM node:22 AS build

WORKDIR /app

COPY package.json package-lock.json next-env.d.ts postcss.config.mjs tsconfig.json next.config.ts ./

RUN npm install

COPY app app
COPY public public

RUN npx next build

# Stage 2: Serve the application
FROM nginx:alpine

EXPOSE 80

COPY --from=build /app/out /usr/share/nginx/html

