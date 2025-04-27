# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Stage 1: Build the application
FROM node:22@sha256:473b4362b26d05e157f8470a1f0686cab6a62d1bd2e59774079ddf6fecd8e37e AS build

WORKDIR /app

COPY package.json package-lock.json next-env.d.ts postcss.config.mjs tsconfig.json next.config.ts ./

RUN npm install

COPY app app
COPY public public

RUN npx next build

# Stage 2: Serve the application
FROM nginxinc/nginx-unprivileged:alpine-perl@sha256:c090c2f38956934e07c1d30f8f28717910be3f33f1c496edfab152ffa0bc7586

EXPOSE 8080

USER nginx

COPY --from=build /app/out /usr/share/nginx/html
