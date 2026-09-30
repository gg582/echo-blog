#!/bin/sh
# Pull the latest code, rebuild the frontend and backend, and restart the
# service. Run as root from anywhere inside the checkout. Content in data/ is
# never touched.
set -eu

cd "$(dirname "$0")/.."
git pull --ff-only

# Build into temporary locations and swap them in, so the running server
# keeps serving a complete frontend and binary until the restart.
(cd blog-frontend && npm ci && BUILD_PATH=build.new npm run build)
(cd blog-backend && CGO_ENABLED=1 go build -o echo-blog.new .)

rm -rf blog-frontend/build.old
if [ -d blog-frontend/build ]; then mv blog-frontend/build blog-frontend/build.old; fi
mv blog-frontend/build.new blog-frontend/build
mv blog-backend/echo-blog.new blog-backend/echo-blog
rm -rf blog-frontend/build.old

systemctl restart echo-blog
systemctl --no-pager --lines=0 status echo-blog
