.PHONY: swagger test-query backup-grafana restore-grafana restore-grafana-file

swagger:
	sh ./resources/swagger.sh

# Backup Grafana configuration
backup-grafana:
	@mkdir -p backups
	docker cp semki-grafana:/var/lib/grafana/grafana.db ./backups/grafana-$(shell date +%Y%m%d-%H%M%S).db

# Restore Grafana configuration
restore-grafana:
	@echo "Available backups:"
	@ls -la ./backups/*.db
	@echo "\nUsage: make restore-grafana-file FILE=./backups/grafana-TIMESTAMP.db"

restore-grafana-file:
ifndef FILE
	$(error FILE is not set, use: make restore-grafana-file FILE=./backups/grafana.db)
endif

	@echo "Setting temporary permissive rights on backup..."
	chmod 666 $(FILE)

	@echo "Copying backup into Grafana volume..."
	docker cp $(FILE) semki-grafana:/var/lib/grafana/grafana.db

	@echo "Restarting Grafana..."
	docker-compose restart grafana

	@echo "Restore complete. Grafana restarted."

# Loki
test-query:
	@echo "=== Last logs ==="
	@curl -s "http://localhost:3100/loki/api/v1/query_range?query={container=~\"semki-.*\"}&limit=5" | jq .


add-hosts:
	sudo sh -c 'echo "127.0.0.1 semki.local" >> /etc/hosts'
