package getenv

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadEnv(path string, masterkey []byte) error {

	envFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}
	defer envFile.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore empty lines and comments (lines starting with #)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Split the line into key and value (key=value)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {

			key, encryptedValue := parts[0], parts[1]

			// Remove apostrophes if there are any
			if encryptedValue[0] == '"' {
				encryptedValue = encryptedValue[1:]
			}
			if encryptedValue[len(encryptedValue)-1] == '"' {
				encryptedValue = encryptedValue[:len(encryptedValue)-1]
			}

			// Decrypt the value
			decryptedValue, err := decrypt(masterkey, encryptedValue)
			if err != nil {
				return fmt.Errorf("error decrypting value for %s: %w", key, err)
			}

			// Set the environment variable using os.Setenv
			err = os.Setenv(key, decryptedValue)
			if err != nil {
				return fmt.Errorf("error setting environment variable %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

func (f *EnvFile) WriteToFile(key, value string) error {

	encryptedValue, err := encrypt(f.key, value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	_, err = fmt.Fprintf(f.file, "%s=%s\n", key, encryptedValue)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	f.file.Sync()

	return nil
}

// encrypt encrypts a string using AES GCM
func encrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(key []byte, encryptedText string) (string, error) {

	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
