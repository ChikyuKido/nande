# Installation with docker
This setup takes a few steps to complete:
#### Step 1
Docker compose. 
```yaml
services:
  nande:
    image: ghcr.io/chikyukido/nande:latest
    container_name: nande
    privileged: true
    environment:
      WEB_PORT: 6643
      EXTENSION_FOLDER: extensions
      INFLUX_URL: http://influxdb:8086
      INFLUX_TOKEN: superscecrettoken # change me to the same as the influx dbs
      INFLUX_ORG: nande
      INFLUX_BUCKET: nande
      GRAFANA_TOKEN: your_grafana_token # the service user api key
      GRAFANA_URL: http://grafana:3000
      GRAFANA_INFLUX_DATASOURCE_ID: your_grafana_influx_datasource_id # the datasource id for the influxdb
    networks:
      nande-network:
    devices:
      - "/dev/sda:/dev/sda" # add your drives here for the smartctl extension. If you want to use it
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./dev-volume:/app/extensions # mount for the extension folder. There you can modify the env variables or add extensions
      - /media/HDD:/mnt/disk1  # Here enter the disk for the smartctl extension. If you want to use it
      - /media/Games:/mnt/disk2
    restart: always
  influxdb:
    image: influxdb:latest
    container_name: influxdb
    environment:
      - INFLUXDB_DB=nande
      - INFLUXDB_HTTP_BIND_ADDRESS=:8086
      - DOCKER_INFLUXDB_INIT_MODE=setup
      - DOCKER_INFLUXDB_INIT_USERNAME=admin
      - DOCKER_INFLUXDB_INIT_PASSWORD=Password
      - DOCKER_INFLUXDB_INIT_ORG=nande
      - DOCKER_INFLUXDB_INIT_BUCKET=nande
      - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=superscecrettoken # change me
    networks:
      nande-network:
    volumes:
      - influxdb-data:/var/lib/influxdb
    restart: always

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    user: '0'
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
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
The variables **GRAFANA_TOKEN** and **GRAFANA_INFLUX_DATASOURCE_ID** can you leave like they are for the moment. <br>
Important is that you put the extension folder somewhere you can access it easily. You have to configure stuff in there.
#### Step 2
Start the docker containers and then go to ip:3000 to access the grafana dashboard. The default credentials are admin/admin <br>
You should change the admin password in the next step. 
After you successfully entered the grafana dashboard go to connection and data sources and add a new source <br>
Choose influxdb <br>
Important is that you choose Flux for the Query language <br>
Then you enter the URL under HTTP as http://influxdb:8086 <br>
Then scroll down to InfluxDB Details and there you change it to the env variable you set in the docker: <br>
If you didnt change anything it is: <br>
Organisation: nande <br>
Token: supersecrettoken (Please change this)<br>
Default Bucket: nande <br>

Then press save & test. If it was successful extract the datasource id from the url at the end. For example: <br>
http://localhost:3000/connections/datasources/edit/degxwphogd62oa <br>
Here is the id degxwphogd62oa. Take this id and put it into the docker compose as **GRAFANA_INFLUX_DATASOURCE_ID**

#### Step 3
Now we need a grafana token. For this we need to go Administration -> Users and access -> Service accounts <br>
In there we add a new service account. The role should be admin. <br>
After you created the service account create a new service account token <br>
This token you paste into the docker compose as **GRAFANA_TOKEN**

#### Step 4
So now we finished the docker compose. Next step is that you recreate the docker compose.
#### Step 5
So now that the container knows the credentials for the grafana and the datasource we can create the dashboards <br>
``docker exec -it nande ./nande grafana create docker``
This creates the docker dashboard. If you want to create dashboard for other extension you just change docker to another extension name
#### Step 6
Go back to grafana and go to dashboards. <br>
Then you should see the docker dashboard. For the docker dashboard you have to change the bucket to nande

# Extension specific setup
## Smartctl
If you want to use the smartctl extension you have to add the devices and mountpoints to the docker compose. <br>
After that you have to go to the extension folder and change the .env file. <br>
You see a variable called SCAN_HDDS. Here you add your drives seperated by a "," in the following format: <br>
drive:mountpoint:price <br>
For example <br>
/dev/sda:/mnt/disk1:200,/dev/sdb:/mnt/disk2:200 <br>
After you changed this variable you need to restart the container. <br> 
To add the smartctl dashboard you call <br>
``docker exec -it nande ./nande grafana create smartctl``



