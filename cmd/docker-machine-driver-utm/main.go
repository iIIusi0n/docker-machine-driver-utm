package main

import (
	"docker-machine-driver-utm/internal/driver"

	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("default", "path"))
}
