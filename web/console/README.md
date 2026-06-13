# web-console

Next.js 14 App Router admin and branch operations console.

**Port**: 3000  
**Auth**: Bearer (access token in memory) + httpOnly refresh cookie  
**Roles**: SUPER_ADMIN (analytics, risk, config, audit), BRANCH_OFFICER (settlements, KYC review, disputes)

## Local run
```bash
cd web/console
npm install
NEXT_PUBLIC_GATEWAY_URL=http://localhost:8000 npm run dev
```
