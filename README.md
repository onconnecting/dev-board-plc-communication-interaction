# Onconnecting Devboard

## Getting Started

- Open a terminal in the mqtt-broker folder
- run `docker compose up -d`

## Folder structure

- PLC-DemoBoard: TIA-Portal V18 Program running on the SPS
- mqtt-broker: docker-compose file for the mqtt-broker and the mqtt data service
- sps-dev-board-mqtt-data-service: MQTT data service written in go. It translates sensor values for the SPS and provied a graphical user interface http://192.168.4.241:8080/ to simulate the sensors
