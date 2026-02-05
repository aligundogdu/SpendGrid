package filesystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"spendgrid/internal/i18n"
)

// Init initializes SpendGrid in the current directory
func Init() error {
	// Check if already initialized
	if _, err := os.Stat(".spendgrid"); err == nil {
		return fmt.Errorf(i18n.T("commands.init.already_exists"))
	}

	// Ask for confirmation
	fmt.Print(i18n.T("commands.init.confirm") + " ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	// Check response - accept both language-specific and universal forms
	yes := i18n.T("common.yes")
	if response != yes && response != "y" && response != "Y" {
		fmt.Println(i18n.T("common.cancel"))
		return nil
	}

	// Create directory structure
	if err := createDirectoryStructure(); err != nil {
		return fmt.Errorf(i18n.Tfmt("commands.init.error", err))
	}

	fmt.Println(i18n.T("commands.init.success"))
	return nil
}

func createDirectoryStructure() error {
	// Create _config directory
	if err := os.MkdirAll("_config", 0755); err != nil {
		return fmt.Errorf("failed to create _config: %v", err)
	}

	// Create _pool directory
	if err := os.MkdirAll("_pool", 0755); err != nil {
		return fmt.Errorf("failed to create _pool: %v", err)
	}

	// Create _share directory
	if err := os.MkdirAll("_share", 0755); err != nil {
		return fmt.Errorf("failed to create _share: %v", err)
	}

	// Create current year directory
	currentYear := strconv.Itoa(time.Now().Year())
	if err := os.MkdirAll(currentYear, 0755); err != nil {
		return fmt.Errorf("failed to create year directory: %v", err)
	}

	// Create .spendgrid version file
	// Format: schema_version build_timestamp
	schemaVersion := "1"
	buildTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	versionInfo := fmt.Sprintf("%s %s\n", schemaVersion, buildTimestamp)

	if err := os.WriteFile(".spendgrid", []byte(versionInfo), 0644); err != nil {
		return fmt.Errorf("failed to create .spendgrid: %v", err)
	}

	// Create config files
	if err := createConfigFiles(); err != nil {
		return err
	}

	// Create year files
	if err := createYearFiles(currentYear); err != nil {
		return err
	}

	return nil
}

func createConfigFiles() error {
	// settings.yml
	settings := `# SpendGrid Local Settings
base_currency: TRY
date_format: "DD.MM.YYYY"
`
	if err := os.WriteFile(filepath.Join("_config", "settings.yml"), []byte(settings), 0644); err != nil {
		return fmt.Errorf("failed to create settings.yml: %v", err)
	}

	// rules.yml
	rules := `# SpendGrid Rules
# Otomatik oluşturulacak düzenli gelir/gider kuralları
rules: []
`
	if err := os.WriteFile(filepath.Join("_config", "rules.yml"), []byte(rules), 0644); err != nil {
		return fmt.Errorf("failed to create rules.yml: %v", err)
	}

	// categories.yml
	categories := `# SpendGrid Categories
# # işareti ile başlayan etiketler
#categories: []
`
	if err := os.WriteFile(filepath.Join("_config", "categories.yml"), []byte(categories), 0644); err != nil {
		return fmt.Errorf("failed to create categories.yml: %v", err)
	}

	// projects.yml
	projects := `# SpendGrid Projects
# @ işareti ile başlayan projeler
#projects: []
`
	if err := os.WriteFile(filepath.Join("_config", "projects.yml"), []byte(projects), 0644); err != nil {
		return fmt.Errorf("failed to create projects.yml: %v", err)
	}

	// backlog.md
	backlog := `# Backlog
# Tarihsiz işlemler, beklenen alacaklar, planlanan büyük harcamalar

`
	if err := os.WriteFile(filepath.Join("_pool", "backlog.md"), []byte(backlog), 0644); err != nil {
		return fmt.Errorf("failed to create backlog.md: %v", err)
	}

	return nil
}

func createYearFiles(year string) error {
	months := []string{
		"Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
		"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık",
	}

	for i, month := range months {
		monthNum := fmt.Sprintf("%02d", i+1)
		filename := filepath.Join(year, monthNum+".md")

		content := fmt.Sprintf(`# %s %s

## ROWS

## RULES

`, year, month)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %v", filename, err)
		}
	}

	return nil
}

// CheckSchemaVersion checks the schema version of the current SpendGrid directory
// Returns the schema version and build timestamp, or error if not initialized
func CheckSchemaVersion() (schemaVersion string, buildTimestamp int64, err error) {
	content, err := os.ReadFile(".spendgrid")
	if err != nil {
		return "", 0, fmt.Errorf("not a spendgrid directory")
	}

	parts := strings.Fields(string(content))
	if len(parts) < 1 {
		return "", 0, fmt.Errorf("invalid .spendgrid file format")
	}

	schemaVersion = parts[0]

	if len(parts) >= 2 {
		buildTimestamp, _ = strconv.ParseInt(parts[1], 10, 64)
	}

	return schemaVersion, buildTimestamp, nil
}

// CheckMigrationNeeded checks if a migration is needed based on schema version
// targetSchema: the required schema version
// Returns true if migration is needed, along with current version
func CheckMigrationNeeded(targetSchema string) (needsMigration bool, currentSchema string, err error) {
	currentSchema, _, err = CheckSchemaVersion()
	if err != nil {
		return false, "", err
	}

	// Compare schema versions
	current, _ := strconv.Atoi(currentSchema)
	target, _ := strconv.Atoi(targetSchema)

	return current < target, currentSchema, nil
}
