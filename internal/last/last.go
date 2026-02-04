package last

import (
	"fmt"
	"strings"

	"spendgrid/internal/config"
)

// ShowRecentDirs displays the list of recent SpendGrid directories
func ShowRecentDirs() error {
	store, err := config.GetRecentDirsStore()
	if err != nil {
		return fmt.Errorf("failed to load recent directories: %v", err)
	}

	dirs := store.GetDirectories()
	if len(dirs) == 0 {
		fmt.Println("Henüz kaydedilmiş SpendGrid dizini yok.")
		fmt.Println("Bir SpendGrid dizinine gidip çalışmaya başlayın.")
		return nil
	}

	fmt.Println("\n📁 Son Kullanılan SpendGrid Dizinleri (Son 10)")
	fmt.Println(strings.Repeat("=", 70))

	for i, dir := range dirs {
		fmt.Printf("%d. %s\n", i+1, dir.Name)
		fmt.Printf("   📍 %s\n", dir.Path)
		fmt.Printf("   🕐 Son kullanım: %s\n", dir.LastUsed.Format("02.01.2006 15:04"))
		if i < len(dirs)-1 {
			fmt.Println()
		}
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\n💡 Bir dizine geçmek için: cd <yukarıdaki yol>")
	return nil
}
