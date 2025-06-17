# Food Ordering API

## Prerequisites
- Docker
- Golang

## Setup Steps

1. Copy the configuration file:
   ```bash
   cp config.local.yaml config.yaml
   ```

2. Create the coupons directory structure:
   ```
   coupons/
   ├── couponbase1
   ├── couponbase2
   └── couponbase3
   ```

3. Start the application:
   ```bash
   docker-compose up --build
   ```

## API Documentation
API endpoints and example requests are available in `api.http`