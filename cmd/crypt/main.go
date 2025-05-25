package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/SqiSch/lpic-cli/internal/database"
	"github.com/SqiSch/lpic-cli/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Command-line arguments
	operation := flag.String("operation", "encrypt", "Operation to perform: encrypt or decrypt")
	var inputFile string
	flag.StringVar(&inputFile, "input", "", "Input file path for decryption")
	var inputURL string
	flag.StringVar(&inputURL, "url", "", "Input URL for decryption")
	outputFile := flag.String("output", "output.json", "Output file path")
	flag.Parse()

	// Get AES key from environment variable
	aesKey := os.Getenv("AES_KEY")
	if aesKey == "" {
		log.Fatal("AES_KEY environment variable is not set")
	}

	key := []byte(aesKey)

	if *operation == "encrypt" {
		loadDataAndEncrypt(key, *outputFile)
	} else if *operation == "decrypt" {
		if inputFile == "" && inputURL == "" {
			log.Fatal("Either input file path or input URL must be provided for decryption")
		}
		if inputURL != "" {
			decryptDataFromURL(key, inputURL, *outputFile)
		} else {
			decryptData(key, inputFile, *outputFile)
		}
	} else {
		log.Fatalf("Invalid operation: %s", *operation)
	}
}

// It prepends the nonce to the ciphertext.
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("failed to create AES cipher block: %w", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("failed to create GCM cipher: %w", err)
		return nil, err
	}

	// GCM requires a unique nonce. We generate a random one for each encryption.
	// The nonce size is determined by the GCM instance.
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("failed to generate nonce: %w", err)
		return nil, err
	}

	// Seal encrypts and authenticates the plaintext.
	// It appends the authentication tag to the ciphertext.
	// The nonce is prepended to the result for easy storage/retrieval.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil) // The first 'nonce' here is the destination buffer - we want the nonce prepended
	return ciphertext, nil
}

// decrypt decrypts the given byte slice (nonce + ciphertext) using AES-GCM.
func decrypt(ciphertextWithNonce []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("failed to create AES cipher block: %w", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("failed to create GCM cipher: %w", err)
		return nil, err
	}

	// Extract the nonce from the beginning of the ciphertextWithNonce
	nonceSize := gcm.NonceSize()
	if len(ciphertextWithNonce) < nonceSize {
		log.Printf("ciphertext too short to contain nonce")
		return nil, err
	}
	nonce := ciphertextWithNonce[:nonceSize]
	ciphertext := ciphertextWithNonce[nonceSize:]

	// Open decrypts and authenticates the data.
	// It verifies the authentication tag appended by Seal().
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil) // The first 'nil' means create a new buffer for plaintext
	if err != nil {
		log.Printf("failed to decrypt and authenticate: %w", err) // This error likely indicates tampering
		return nil, err
	}

	return plaintext, nil
}

func loadDataAndEncrypt(key []byte, outputFile string) {
	certIds := []string{"lpic1-101-500", "lpic1-102-500", "lpic2-202-450", "lpic2-201-450", "cka"}

	mongoClient := database.MongoConnect()

	// Fetch the CertificationSet from MongoDB
	collection := mongoClient.Database("certificationDB").Collection("questions")
	var certificationSets []types.CertificationSet

	for _, certID := range certIds {
		var certificationSet types.CertificationSet

		fmt.Println("Fetching CertificationSet with ID: ", certID)
		err := collection.FindOne(context.TODO(), bson.M{"_id": certID}).Decode(&certificationSet)
		if err != nil {
			log.Fatalf("Failed to fetch CertificationSet: %v", err)
		}

		certificationSets = append(certificationSets, certificationSet)
	}

	// Marshal the data
	data, err := json.Marshal(certificationSets)
	if err != nil {
		log.Fatalf("Failed to marshal data: %v", err)
	}

	// Compress the data
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	_, err = gzipWriter.Write(data)
	if err != nil {
		log.Fatalf("Failed to compress data: %v", err)
	}
	gzipWriter.Close()

	// Encrypt the compressed data
	encryptedData, err := encrypt(compressedData.Bytes(), key)
	if err != nil {
		log.Fatalf("Failed to encrypt data: %v", err)
	}

	// Write encrypted data to file
	err = ioutil.WriteFile(outputFile, encryptedData, 0644)
	if err != nil {
		log.Fatalf("Failed to write encrypted data to file: %v", err)
	}

	log.Printf("Data successfully compressed, encrypted, and saved to %s", outputFile)
}

func decryptData(key []byte, inputFile string, outputFile string) {

	// Read encrypted data from decompressed file
	encryptedData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read encrypted data from file: %v", err)
	}

	// Decrypt the data
	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		log.Fatalf("Failed to decrypt data: %v", err)
	}

	// Decompress the decrypted data
	var decompressedData bytes.Buffer
	gzipReader, err := gzip.NewReader(bytes.NewReader(decryptedData))
	if err != nil {
		log.Fatalf("Failed to decompress data: %v", err)
	}

	_, err = io.Copy(&decompressedData, gzipReader)
	gzipReader.Close()
	if err != nil {
		log.Fatalf("Failed to read decompressed data: %v", err)
	}

	// Unmarshal the decompressedData data
	var certificationSets []types.CertificationSet
	err = json.Unmarshal(decompressedData.Bytes(), &certificationSets)
	if err != nil {
		log.Fatalf("Failed to unmarshal decrypted data: %v", err)
	}

	// Write decrypted data to output file
	decryptedJSON, err := json.MarshalIndent(certificationSets, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal decrypted data: %v", err)
	}

	err = ioutil.WriteFile(outputFile, decryptedJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write decrypted data to file: %v", err)
	}

	log.Printf("Data successfully decrypted and saved to %s", outputFile)
}

func decryptDataFromURL(key []byte, inputURL string, outputFile string) {
	// Fetch encrypted data from URL
	resp, err := http.Get(inputURL)
	if err != nil {
		log.Fatalf("Failed to fetch encrypted data from URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch encrypted data: HTTP %d", resp.StatusCode)
	}

	encryptedData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read encrypted data from response: %v", err)
	}

	// Decrypt the encrypted data
	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		log.Fatalf("Failed to decrypt data: %v", err)
	}

	// Decompress the decrypted data
	var decompressedData bytes.Buffer
	gzipReader, err := gzip.NewReader(bytes.NewReader(decryptedData))
	if err != nil {
		log.Fatalf("Failed to decompress data: %v", err)
	}
	_, err = io.Copy(&decompressedData, gzipReader)
	gzipReader.Close()
	if err != nil {
		log.Fatalf("Failed to read decompressed data: %v", err)
	}

	// Write decrypted data to output file
	err = ioutil.WriteFile(outputFile, decompressedData.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Failed to write decrypted data to file: %v", err)
	}

	log.Printf("Data successfully decrypted and saved to %s", outputFile)
}
