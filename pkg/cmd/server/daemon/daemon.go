package daemon

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/server/logs"
	"github.com/kardianos/service"
)

type program struct {
	service.Interface
}

func Start() error {
	cfg, err := getServiceConfig()
	if err != nil {
		return err
	}
	s, err := service.New(program{}, cfg)
	if err != nil {
		return err
	}
	err = s.Install()
	if err != nil {
		return err
	}

	err = s.Start()
	if err != nil {
		return err
	}

	logFilePath := logs.GetLogFilePath()

	time.Sleep(5 * time.Second)
	status, err := s.Status()
	if err != nil {
		return err
	}

	if status == service.StatusRunning {
		return nil
	}

	err = Stop()
	if err != nil {
		return err
	}
	logContent, err := os.ReadFile(*logFilePath)
	if err != nil {
		return err
	}
	fmt.Println(string(logContent))
	if status == service.StatusStopped {
		return fmt.Errorf("daemon stopped unexpectedly")
	} else {
		return fmt.Errorf("daemon status unknown")
	}
}

func Stop() error {
	cfg, err := getServiceConfig()
	if err != nil {
		return err
	}
	s, err := service.New(program{}, cfg)
	if err != nil {
		return err
	}

	err = s.Stop()
	if err != nil {
		return err
	}

	return s.Uninstall()
}

func getServiceConfig() (*service.Config, error) {
	svcConfig := &service.Config{
		Name:        "DaytonaServerDaemon",
		DisplayName: "Daytona Server",
		Description: "This is the Daytona Server daemon.",
		Arguments:   []string{"server"},
	}

	user := os.Getenv("USER")

	switch runtime.GOOS {
	case "windows":
		return nil, fmt.Errorf("daemon mode not supported on Windows")
	case "linux":
		if !strings.HasSuffix(service.Platform(), "systemd") {
			return nil, fmt.Errorf("on Linux, `server -d` is only supported with systemd. %s detected", service.Platform())
		}
		fallthrough
	case "darwin":
		if user != "" && user != "root" {
			svcConfig.Option = service.KeyValue{"UserService": true}
		}
	}

	return svcConfig, nil
}
