# Installation with docker
Docker compose. 
```yaml
services:
  nande:
    image: ghcr.io/chikyukido/nande:latest
    container_name: nande
    environment:
      WEB_PORT: 6643
      EXTENSION_FOLDER: extension-build
      INFLUX_URL: http://influxdb:8086
      INFLUX_TOKEN: superscecrettoken # change this to the same as in the influxdb
      INFLUX_ORG: nande
      INFLUX_BUCKET: nande
      GRAFANA_TOKEN: your_grafana_token # create a service user in grafana
      GRAFANA_URL: http://grafana:3000
      GRAFANA_INFLUX_DATASOURCE_ID: your_grafana_influx_datasource_id
    ports:
      - "6643:6643"
    networks:
      nande-network:
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./dev-volume:/app/extensions # mount for the extension folder. There you can modify the env variables or add extensions
    restart: always
  influxdb:
    image: influxdb:latest
    container_name: influxdb
    environment:
      - INFLUXDB_DB=nande
      - INFLUXDB_HTTP_BIND_ADDRESS=:8086
      - DOCKER_INFLUXDB_INIT_MODE=setup
      - DOCKER_INFLUXDB_INIT_USERNAME=admin
      - DOCKER_INFLUXDB_INIT_PASSWORD=Password # change the default password
      - DOCKER_INFLUXDB_INIT_ORG=nande
      - DOCKER_INFLUXDB_INIT_BUCKET=nande
      - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=superscecrettoken # change me
    ports:
      - "8086:8086"
    networks:
      nande-network:
    volumes:
      - influxdb-data:/var/lib/influxdb
    restart: always

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    ports:
      - "3000:3000"
    networks:
      nande-network:
    volumes:
      - grafana-data:/var/lib/grafana
    restart: always

networks:
  nande-network:
volumes:
  influxdb-data:
  grafana-data:
```