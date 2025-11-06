.PHONY: dev run

dev:
	cd client && npm i && npm run dev &
	cd server && docker compose up -d --build

run:
	cd client && docker compose up -d
	cd server && docker compose up -d