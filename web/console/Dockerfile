FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY .next/standalone ./
COPY .next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
