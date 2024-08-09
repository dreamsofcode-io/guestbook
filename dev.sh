#!/usr/bin/env bash

air -c ./.air.toml &
tailwindcss \
  -i 'static/css/main.css' \
  -o 'static/css/style.css' \
  --watch & \
browser-sync start \
  --files 'templates/**/*.html, static/**/*.css' \
  --port 3001 \
  --proxy 'localhost:8080' \
  --middleware 'function(req, res, next) { \
    res.setHeader("Cache-Control", "no-cache, no-store, must-revalidate"); \
    return next(); \
  }'
 #& \
#templ generate -watch
