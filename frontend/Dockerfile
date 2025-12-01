# ============================================
# Build Stage
# ============================================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files (cached layer)
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build for production
RUN npm run build

# ============================================
# Production Stage (Nginx)
# ============================================
FROM nginx:1.25-alpine

# Install curl for health checks
RUN apk --no-cache add curl

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# Create non-root user
RUN addgroup -g 1000 nginx-user && \
    adduser -D -u 1000 -G nginx-user nginx-user && \
    chown -R nginx-user:nginx-user /usr/share/nginx/html && \
    chown -R nginx-user:nginx-user /var/cache/nginx && \
    touch /var/run/nginx.pid && \
    chown nginx-user:nginx-user /var/run/nginx.pid

# Switch to non-root user
USER nginx-user

EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]

