# Docker Machine UTM Driver

A Docker Machine driver for [UTM](https://mac.getutm.app/) virtual machines on macOS. This driver allows you to create and manage Docker machines running inside UTM VMs.

## Requirements

- UTM 4.2 or later

## Installation

1. Download the latest release from the releases page or build from source:
    ```bash
    make build
    ```

2. Copy the binary to your PATH:
    ```bash
    sudo cp bin/docker-machine-driver-utm /usr/local/bin
    ```


## Usage

Create a new Docker machine with default settings:

```bash
docker-machine create --driver utm manager
```


Available driver options:

- `--utm-memory`: Memory size for VM in MB (default: 1024)
- `--utm-disk`: Disk size for VM in MB (default: 8192)
- `--utm-cpu`: Number of CPU cores (default: 1)
- `--utm-network`: Network type (emulated, shared, host, bridged) (default: shared)
- `--utm-host-interface`: Host interface for bridged networking
- `--utm-boot2docker-url`: Custom URL for boot2docker ISO
- `--utm-ssh-user`: SSH username (default: docker)

Example with custom settings:
```bash
docker-machine create --driver utm \
--utm-memory 2048 \
--utm-disk 20000 \
--utm-cpu 2 \
--utm-network bridged \
--utm-host-interface en0 \
my-docker-vm
```


## Notes

- This driver uses a custom boot2docker ISO with QEMU guest agent support for better integration with UTM. The ISO will be updated in future releases.
- The driver supports all standard Docker Machine commands (start, stop, restart, rm, etc.)
- Network modes:
  - `shared`: Uses UTM's shared network (recommended)
  - `bridged`: Connects directly to host network interface
  - `host`: Host-only networking
  - `emulated`: Emulated network device

