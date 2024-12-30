package getenv

import "os"

type EnvFile struct {
	file *os.File
	key  []byte
}
