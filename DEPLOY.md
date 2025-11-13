# Deploying Leep Audio Backend to Render.com

## Prerequisites
- GitHub repository with your code pushed
- Render.com account (free tier works)
- Supabase credentials ready

---

## Step 1: Push Code to GitHub

```bash
# Make sure all your code is committed
git add .
git commit -m "Complete Week 2 & 3 backend with Supabase integration"
git push origin main
```

---

## Step 2: Create New Web Service on Render

1. Go to https://render.com/dashboard
2. Click **"New +"**  **"Web Service"**
3. Connect your GitHub repository
4. Select the `Leep_Backend` repository

---

## Step 3: Configure Build Settings

### Basic Settings:
- **Name**: `leep-audio-backend`
- **Region**: Choose closest to your users (e.g., Oregon USA)
- **Branch**: `main`
- **Root Directory**: Leave blank (or specify if nested)
- **Runtime**: `Go`

### Build Command:
```bash
go build -o leep_backend main.go
```

### Start Command:
```bash
./leep_backend
```

---

## Step 4: Set Environment Variables

Click **"Advanced"** and add these environment variables:

```bash
PORT=3000
NODE_ENV=production

# Supabase Configuration
SUPABASE_URL=https://xblkxhfqwgvhgiginmbl.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InhibGt4aGZxd2d2aGdpZ2lubWJsIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjE2NTA5NDcsImV4cCI6MjA3NzIyNjk0N30.pfAUiYh-rKlHVaohsAeypVbJqF_r_ItGjJjrB0fBsvo
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InhibGt4aGZxd2d2aGdpZ2lubWJsIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2MTY1MDk0NywiZXhwIjoyMDc3MjI2OTQ3fQ.OCrk5OxQ3LR3K5lN9yu087MHbu4IRR3GD3S830VtWWQ
SUPABASE_JWT_SECRET=HXc2jPoHyFYIQpbDj4v467XwZ/xICe460toFIpxZSLnIpsV3G84HBqXHF1PQawQ6hBVEFUTp5hlgCPmXK0Mw9Q==

JWT_EXPIRY=3600
```

** SECURITY NOTE**: In production, rotate these keys regularly and never commit them to git!

---

## Step 5: Configure Instance Type

- **Instance Type**: Free tier (or upgrade for better performance)
- **Auto-Deploy**: Yes (deploys on every git push)

---

## Step 6: Deploy!

1. Click **"Create Web Service"**
2. Render will start building and deploying
3. Watch the logs for any errors
4. Deployment takes 2-5 minutes

---

## Step 7: Test Your Deployment

Once deployed, you'll get a URL like: `https://leep-audio-backend.onrender.com`

### Test the health endpoint:
```bash
curl https://leep-audio-backend.onrender.com/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "leep-backend",
  "time": "2025-01-12T..."
}
```

### Test the status endpoint:
```bash
curl https://leep-audio-backend.onrender.com/api/v1/status
```

---

## Step 8: Update Frontend Environment Variables

Update your Vercel frontend with the new backend URL:

```bash
# In your frontend .env
NEXT_PUBLIC_API_URL=https://leep-audio-backend.onrender.com
```

---

## Troubleshooting

### Build Fails
- Check Render logs for specific error
- Ensure `go.mod` and `go.sum` are committed
- Verify Go version compatibility

### App Crashes on Startup
- Check environment variables are set correctly
- Review startup logs for missing dependencies
- Verify Supabase credentials are correct

### 401 Errors
- Check JWT secret is correctly encoded
- Verify Supabase anon key and service role key
- Check token expiration time

### CORS Errors
- Update `internal/middleware/cors.go` to restrict origins in production
- Change `Access-Control-Allow-Origin` from `*` to your Vercel domain:
  ```go
  c.Writer.Header().Set("Access-Control-Allow-Origin", "https://your-app.vercel.app")
  ```

---

## Monitoring & Logs

### View Logs:
1. Go to your Render dashboard
2. Click on your service
3. Click **"Logs"** tab
4. Logs show requests with user tracking

### Metrics:
- Render provides basic metrics (CPU, Memory, Bandwidth)
- For advanced monitoring, integrate tools like:
  - Sentry (error tracking)
  - Prometheus (metrics)
  - LogDNA (log aggregation)

---

## Auto-Deploy on Git Push

Once configured, every push to `main` branch will trigger a new deployment:

```bash
git add .
git commit -m "Add new feature"
git push origin main
# Render automatically deploys!
```

---

## Upgrading

### From Free to Paid Tier:
1. Go to Render dashboard
2. Click on your service
3. Click **"Upgrade"**
4. Choose Starter ($7/month) or higher

Benefits:
- No cold starts
- More resources
- Better performance
- Custom domains

---

## Custom Domain (Optional)

1. Go to service settings
2. Click **"Custom Domain"**
3. Add your domain (e.g., `api.leepaudio.com`)
4. Update DNS records as instructed
5. Render provides free SSL certificate

---

## Production Checklist

- [ ] Environment variables set correctly
- [ ] CORS restricted to frontend domain
- [ ] Rate limiting configured appropriately
- [ ] Error tracking set up (Sentry recommended)
- [ ] Monitoring configured
- [ ] Custom domain configured (optional)
- [ ] SSL certificate enabled (automatic on Render)
- [ ] Auto-deploy enabled for main branch
- [ ] Team notified of deployment
- [ ] API documentation shared with frontend team

---

## Next Steps

1. Test all API endpoints in production
2. Share API URL with frontend team
3. Set up monitoring and alerting
4. Plan for scaling as usage grows

---

## Support

- **Render Docs**: https://render.com/docs
- **Render Community**: https://community.render.com
- **Supabase Docs**: https://supabase.com/docs

Good luck with your deployment! 
