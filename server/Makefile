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
	docker cp $(FILE) semki-grafana:/var/lib/grafana/grafana.db
	docker-compose restart grafana

# Loki
test-query:
	@echo "=== Last logs ==="
	@curl -s "http://localhost:3100/loki/api/v1/query_range?query={container=~\"semki-.*\"}&limit=5" | jq .
