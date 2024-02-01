package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/kirill-shtrykov/yaml"
)

type Secret struct {
	path   string
	fields map[string]interface{}
}

type Secrets []*Secret

func (s *Secrets) Add(path string, field map[string]interface{}) {
	for _, secret := range *s {
		if secret.path == path {
			for k, v := range field {
				secret.fields[k] = v
			}
			return
		}
	}
	secret := Secret{path: path, fields: field}
	*s = append(*s, &secret)
}

// Fill secrets from YAML map
func (s *Secrets) Fill(m map[string]interface{}) {
	var walk func(path string, m map[string]interface{})
	walk = func(path string, m map[string]interface{}) {
		for k, v := range m {
			if _, ok := v.(map[string]interface{}); ok {
				walk(path+"/"+k, v.(map[string]interface{}))
			} else {
				s.Add(path, map[string]interface{}{k: v})
			}
		}
	}

	walk("", m)
}

// Retrieves the value of the environment variable named by the `key`
// It returns the value if variable present and value not empty
// Otherwice it returns string value `def`
func stringFromEnv(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimSpace(v)
	}
	return def
}

// Reads the named file and returns the contents as string
// If file reading attempt is unsuccsessful it returns string value `def`
func stringFromFile(name string, def string) string {
	c, err := os.ReadFile(expandUserHomeDir(name))
	if err != nil {
		return def
	}
	return strings.TrimSpace(string(c))
}

func getUserHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

// Expands linux home dir (~) to full path
func expandUserHomeDir(path string) string {
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		return getUserHomeDir()
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		return filepath.Join(getUserHomeDir(), path[2:])
	}
	return path
}

// Reads YAML file to map
func readYAML(path string) map[string]interface{} {
	var data map[string]interface{}

	fp, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Load(fp, &data, nil); err != nil {
		log.Fatal("failed to parse YAML job: %w", err)
	}

	return data
}

// Create Vault client
func initVaultClient() *vault.Client {
	config := vault.DefaultConfig()
	if err := config.ReadEnvironment(); err != nil {
		log.Fatal("vault config: Failed to read environment varables: %w", err)
	}

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatal("unable to initialize Vault client: %w", err)
	}

	return client
}

func run() error {
	var (
		secretFile string // YAML file with secrets
		mountPath  string // The path to the KV mount to config, such as secret. This is specified as part of the URL
	)

	flag.StringVar(&secretFile, "secrets", stringFromEnv("SECRETS", "./secrets.yaml"), "YAML file with secrets")
	flag.StringVar(&mountPath, "mount", stringFromEnv("MOUNT", "secret"), "The path to the KV mount")

	client := initVaultClient()
	yamlData := readYAML(secretFile)

	var secrets Secrets
	secrets.Fill(yamlData)
	for _, s := range secrets {
		log.Printf("%s: %v", s.path, s.fields)
		_, err := client.KVv2(mountPath).Put(context.Background(), s.path, s.fields)
		if err != nil {
			log.Fatal()
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
