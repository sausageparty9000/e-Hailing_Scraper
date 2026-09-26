FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:24-alpine
WORKDIR /app
COPY --from=build /app /app
EXPOSE 80
CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0", "--port", "80"]
