FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install
RUN npm install recharts

COPY . .

EXPOSE 3000
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000", "--no-open"] 