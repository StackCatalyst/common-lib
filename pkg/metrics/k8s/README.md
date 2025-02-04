# Prometheus Service Discovery Configuration

This directory contains Kubernetes configurations for setting up Prometheus service discovery with the TerraOrbit services.

## Components

1. `servicemonitor.yaml`: Defines how Prometheus should discover and scrape metrics from TerraOrbit services
   - Targets services with label `app: terraorbit`
   - Scrapes metrics every 15 seconds
   - Includes sample and target limits for resource management

2. `service.yaml`: Template for exposing metrics endpoints
   - Creates a dedicated metrics service
   - Uses standard Prometheus annotations
   - Exposes metrics on port 9090

## Usage

1. Apply the ServiceMonitor:
   ```bash
   kubectl apply -f servicemonitor.yaml
   ```

2. For each service, create a metrics service:
   ```bash
   # Replace values in service.yaml and apply
   kubectl apply -f service.yaml
   ```

3. Ensure your service pods have the following:
   - Label: `app: terraorbit`
   - Port named `metrics` exposing :9090

## Configuration

The default configuration assumes:
- Prometheus operator is installed
- Services are in the `terraorbit` namespace
- Metrics are exposed on port 9090 at `/metrics`

Adjust the configurations as needed for your environment. 