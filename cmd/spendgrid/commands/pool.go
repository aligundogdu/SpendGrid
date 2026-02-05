package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/pool"
)

var PoolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Manage pool/backlog items",
	Long:  `Manage items in the pool/backlog. Pool items are future expenses that haven't been assigned to a specific month yet.`,
}

var PoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pool items",
	Long:  `Display all items currently in the pool/backlog.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := pool.ShowPool(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}

var PoolAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add item to pool",
	Long:  `Add a new expense item to the pool/backlog for future use.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := pool.AddPoolItem(); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("Item added to pool successfully")
	},
}

var PoolMoveCmd = &cobra.Command{
	Use:   "move <line-number> <month>",
	Short: "Move item from pool to month",
	Long:  `Move a pool item to a specific month (1-12).`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lineNum := args[0]
		month := args[1]
		if err := pool.MovePoolItem(lineNum, month); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("Item moved successfully")
	},
}

var PoolRemoveCmd = &cobra.Command{
	Use:   "remove <line-number>",
	Short: "Remove item from pool",
	Long:  `Remove an item from the pool/backlog by its line number.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lineNum := args[0]
		if err := pool.RemovePoolItem(lineNum); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("Item removed from pool successfully")
	},
}

func init() {
	PoolCmd.AddCommand(PoolListCmd)
	PoolCmd.AddCommand(PoolAddCmd)
	PoolCmd.AddCommand(PoolMoveCmd)
	PoolCmd.AddCommand(PoolRemoveCmd)
}
