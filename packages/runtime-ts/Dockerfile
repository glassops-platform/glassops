# GlassOps Runtime - Docker Container Action
FROM node:20-alpine

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    bash \
    ca-certificates \
    openssh-client \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy package files first for better layer caching
COPY package*.json ./
COPY tsconfig.json ./

# Install dependencies
RUN npm ci --production=false

# Copy source code
COPY src/ ./src/

# Build TypeScript
RUN npm run build

# Create non-root user for security
RUN addgroup -S glassops && adduser -S glassops -G glassops
RUN chown -R glassops:glassops /app

# Switch to non-root user
USER glassops

# Set environment variables
ENV NODE_ENV=production
ENV PATH="/app/node_modules/.bin:${PATH}"

# Entrypoint
ENTRYPOINT ["node", "/app/dist/index.js"]