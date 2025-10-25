package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lich0821/ccNexus/internal/config"
)

func main() {
	// Test 1: Create a default config
	fmt.Println("Test 1: Creating default config...")
	defaultConfig := config.DefaultConfig()
	
	// Test GetAnthropicModel
	model, err := defaultConfig.GetAnthropicModel("Claude Official")
	if err != nil {
		log.Fatalf("Failed to get AnthropicModel: %v", err)
	}
	fmt.Printf("Default AnthropicModel: %s\n", model)
	
	// Test 2: Update AnthropicModel
	fmt.Println("\nTest 2: Updating AnthropicModel...")
	err = defaultConfig.UpdateAnthropicModel("Claude Official", "claude-3-opus-20240229")
	if err != nil {
		log.Fatalf("Failed to update AnthropicModel: %v", err)
	}
	
	// Verify the update
	model, err = defaultConfig.GetAnthropicModel("Claude Official")
	if err != nil {
		log.Fatalf("Failed to get updated AnthropicModel: %v", err)
	}
	fmt.Printf("Updated AnthropicModel: %s\n", model)
	
	// Test 3: Save config to file
	fmt.Println("\nTest 3: Saving config to file...")
	configPath := "test_config.json"
	err = defaultConfig.Save(configPath)
	if err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}
	fmt.Printf("Config saved to %s\n", configPath)
	
	// Test 4: Load config from file
	fmt.Println("\nTest 4: Loading config from file...")
	loadedConfig, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify the loaded config
	model, err = loadedConfig.GetAnthropicModel("Claude Official")
	if err != nil {
		log.Fatalf("Failed to get AnthropicModel from loaded config: %v", err)
	}
	fmt.Printf("AnthropicModel from loaded config: %s\n", model)
	
	// Test 5: Test error case - non-existent endpoint
	fmt.Println("\nTest 5: Testing error case - non-existent endpoint...")
	_, err = defaultConfig.GetAnthropicModel("Non-existent Endpoint")
	if err != nil {
		fmt.Printf("Expected error for non-existent endpoint: %v\n", err)
	} else {
		fmt.Println("Error: Expected error for non-existent endpoint but got none")
	}
	
	// Clean up
	fmt.Println("\nCleaning up test file...")
	err = os.Remove(configPath)
	if err != nil {
		log.Printf("Failed to remove test file: %v", err)
	}
	
	fmt.Println("\nAll tests completed successfully!")
}