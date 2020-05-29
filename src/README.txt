# Local development
This will start a local server and can be accessed at http://localhost:8081/
1. Run `npm install`
2. Run `npm run start`

# Deployment
This will auto-deploy website to gh-pages (if pushed to gh-pages branch). If a CNAME is present, it will deploy to that URL also.
1. Run `npm install` (if not already)
2. Run `npm run build`
3. Commit & push built files at root of repository
