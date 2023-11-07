FROM golang:1.21 AS base-go

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


FROM base-go as build-go

COPY . .
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-th


FROM node:20 AS base-node

WORKDIR /app

ENV NODE_ENV="production"
RUN npm i -g pnpm


FROM base-node AS build-node

COPY . .
RUN pnpm i && pnpm build


FROM base-go

COPY --from=build-go /go-th /go-th
COPY --from=build-node /app/build /app/build

EXPOSE 3000
CMD ["/go-th"]
