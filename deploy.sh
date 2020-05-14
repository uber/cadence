#!/usr/bin/env sh

# abort on errors
set -e

# build
npm run docs:build

# navigate into the build output directory
cd docs/.vuepress/dist

# if you are deploying to a custom domain
# echo 'cadenceworkflow.io' > CNAME

git init
git add -A
git commit -m 'deploy'

# deploy to https://uber.github.io/cadence
# will redirect to https://cadenceworkflow.io
git push -f git@github.com:uber/cadence.git master:gh-pages

cd -