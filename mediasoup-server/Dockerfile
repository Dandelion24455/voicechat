FROM node:22-bookworm-slim
WORKDIR /app
COPY mediasoup-server/package.json ./
RUN npm install --omit=dev
COPY mediasoup-server/src/ ./src/
EXPOSE 3000
CMD ["node", "src/server.js"]
