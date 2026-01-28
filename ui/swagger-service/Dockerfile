FROM nginx:1.27-alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static assets (swagger docs + UI)
COPY public /usr/share/nginx/html

# Default port used in docker-compose and Railway
ENV PORT=8080
EXPOSE 8080
