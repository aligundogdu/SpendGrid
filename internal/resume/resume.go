package resume

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"spendgrid/internal/config"
)

// ShowRecentDirs displays the list of recent directories and lets user select one
func ShowRecentDirs() (string, error) {
	store, err := config.GetRecentDirsStore()
	if err != nil {
		return "", fmt.Errorf("failed to load recent directories: %v", err)
	}

	dirs := store.GetDirectories()
	if len(dirs) == 0 {
		fmt.Println("Henüz kaydedilmiş SpendGrid dizini yok.")
		fmt.Println("Bir SpendGrid dizinine gidip çalışmaya başlayın.")
		return "", nil
	}

	fmt.Println("\n📁 Son Kullanılan SpendGrid Dizinleri")
	fmt.Println(strings.Repeat("=", 60))

	for i, dir := range dirs {
		fmt.Printf("%d. %s\n", i+1, dir.Name)
		fmt.Printf("   📍 %s\n", dir.Path)
		fmt.Printf("   🕐 Son kullanım: %s\n", dir.LastUsed.Format("02.01.2006 15:04"))
		if i < len(dirs)-1 {
			fmt.Println()
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Print("\nSeçiminiz (1-10) veya çıkmak için 'q': ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "Q" {
		return "", nil
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(dirs) {
		fmt.Println("❌ Geçersiz seçim!")
		return "", nil
	}

	selectedDir := dirs[choice-1].Path

	// Check if directory still exists
	if _, err := os.Stat(selectedDir); os.IsNotExist(err) {
		fmt.Printf("❌ Dizin bulunamadı: %s\n", selectedDir)
		fmt.Println("Dizin silinmiş veya taşınmış olabilir.")
		return "", nil
	}

	return selectedDir, nil
}

// ListRecentDirs simply lists recent directories without selection
func ListRecentDirs() error {
	store, err := config.GetRecentDirsStore()
	if err != nil {
		return fmt.Errorf("failed to load recent directories: %v", err)
	}

	dirs := store.GetDirectories()
	if len(dirs) == 0 {
		fmt.Println("Henüz kaydedilmiş SpendGrid dizini yok.")
		return nil
	}

	fmt.Println("\n📁 Son Kullanılan SpendGrid Dizinleri")
	fmt.Println(strings.Repeat("=", 60))

	for i, dir := range dirs {
		fmt.Printf("%d. %s\n", i+1, dir.Name)
		fmt.Printf("   📍 %s\n", dir.Path)
		fmt.Printf("   🕐 %s\n", dir.LastUsed.Format("02.01.2006 15:04"))
		if i < len(dirs)-1 {
			fmt.Println()
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	return nil
}
