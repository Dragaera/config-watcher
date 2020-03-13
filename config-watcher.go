package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	TargetFiles               []string       // List of files which to watch
	TargetProcess             string         // Name of process which to reload on changes to files
	ReloadSignal              syscall.Signal // Signal which to send to process
	SleepDuration             int            // Duration in seconds which to sleep between checking files for changes
	SleepBeforeReloadDuration int            // Duration in seconds which to sleep when having detected a change in the config file, before sending the signal to the process
	Verbose                   bool           // Whether to enable verbose logging
}

var (
	logger *log.Logger
)

func main() {
	oldFileHashes := make(map[string]string)
	newFileHashes := make(map[string]string)

	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	config, err := getConfig()
	if err != nil {
		logger.Fatalln("Fatal:", err)
	}

	// Initialize hashes
	updateFileHashes(&config, oldFileHashes)
	logger.Println("Initialized hashes:")
	for path, hash := range oldFileHashes {
		logger.Printf("\t%v => %v", path, hash)
	}

	for {
		updateFileHashes(&config, newFileHashes)

		for path, oldHash := range oldFileHashes {
			newHash := newFileHashes[path]

			if oldHash != newHash {
				logger.Printf("File %v changed. Old hash: %v, new hash: %v",
					path, oldHash, newHash)
				err = reloadProcess(&config)
				if err != nil {
					logger.Printf("Unable to reload process: %v", err)
				}
				oldFileHashes[path] = newHash
			}
		}

		time.Sleep(time.Duration(config.SleepDuration) * time.Second)
	}

}

func reloadProcess(config *Config) error {
	// Sleep to ensure change is visible in all containers of the pod.
	time.Sleep(time.Duration(config.SleepBeforeReloadDuration) * time.Second)
	pid, err := getPidByName(config.TargetProcess)
	if err != nil {
		return err
	}

	logger.Printf("Sending signal %v to process %v", config.ReloadSignal, pid)
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Unable to send signal to process: %v", err)
	}

	err = proc.Signal(config.ReloadSignal)
	if err != nil {
		return fmt.Errorf("Unable to send signal to process: %v", err)
	}

	return nil
}

func getPidByName(executable string) (int, error) {
	// Output is ordered by PID, so we can simply pick the first
	procs, err := ps.Processes()
	if err != nil {
		return -1, fmt.Errorf("Unable to retrieve process list: %v", err)
	}

	for _, proc := range procs {
		if proc.Executable() == executable {
			pid := proc.Pid()
			return pid, nil
		}
	}

	return -1, fmt.Errorf("Did not find process '%v'", executable)
}

func updateFileHashes(config *Config, hashes map[string]string) {
	for _, path := range config.TargetFiles {
		hash, err := hashFile(path)
		if err != nil {
			logger.Printf("Unable to hash file '%s': %s", path, err)
			hashes[path] = ""
			continue
		}

		hashes[path] = hash
	}
}

func hashFile(path string) (string, error) {
	var hashStr string

	f, err := os.Open(path)
	if err != nil {
		return hashStr, err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return hashStr, err
	}
	hashStr = hex.EncodeToString(hash.Sum(nil))

	return hashStr, nil
}

func getConfig() (Config, error) {
	cfg := Config{}

	// TARGET_FILES
	targetFiles, err := stringFromEnv("TARGET_FILES")
	if err != nil {
		return cfg, err
	}
	cfg.TargetFiles = strings.Split(targetFiles, ",")

	// RELOAD_SIGNAL
	reloadSignalName, err := stringFromEnv("RELOAD_SIGNAL")
	if err != nil {
		return cfg, err
	}
	reloadSignal := unix.SignalNum(reloadSignalName)
	if reloadSignal == 0 {
		return cfg, fmt.Errorf("Unknown signal: %v", reloadSignalName)
	}
	cfg.ReloadSignal = reloadSignal

	// TARGET_PROCESS
	targetProcess, err := stringFromEnv("TARGET_PROCESS")
	if err != nil {
		return cfg, err
	}
	cfg.TargetProcess = targetProcess

	// VERBOSE
	verbose, exists := os.LookupEnv("VERBOSE")
	cfg.Verbose = exists && len(verbose) > 0

	// SLEEP_DURATION
	sleepDuration, err := intFromEnvWithDefault("SLEEP_DURATION", 1)
	if err != nil {
		return cfg, err
	}
	cfg.SleepDuration = sleepDuration

	// SLEEP_BEFORE_RELOAD_DURATION
	sleepBeforeReloadDuration, err := intFromEnvWithDefault("SLEEP_BEFORE_RELOAD_DURATION", 1)
	if err != nil {
		return cfg, err
	}
	cfg.SleepBeforeReloadDuration = sleepBeforeReloadDuration

	return cfg, nil
}

func stringFromEnv(key string) (string, error) {
	val, exists := os.LookupEnv(key)
	if !exists {
		return val, fmt.Errorf("Missing env variable '%v'", key)
	}

	return val, nil
}

func intFromEnvWithDefault(key string, fallback int) (int, error) {
	val, err := stringFromEnv(key)
	if err != nil {
		return fallback, nil
	}
	valNumeric, err := strconv.Atoi(val)
	if err != nil {
		return valNumeric, fmt.Errorf("Could not convert to integer: '%v'", val)
	}

	return valNumeric, nil
}
