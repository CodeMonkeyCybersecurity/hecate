// main.go

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"hecate/cmd"
	"hecate/pkg/logger"
	"go.uber.org/zap" 
)


func main() {
    // Initialize logging
    logger.Initialize()
    log := logger.GetSafeLogger()
    defer logger.Sync() // ✅ Ensures logs are flushed properly
	
    fallbackLogging := log == nil

    if fallbackLogging {
        println("⚠️ Warning: Logger is nil. Defaulting to console output.")
    }

    // Register all commands in one place
    cmd.RegisterCommands()

    // Execute the root command
    if err := cmd.RootCmd.Execute(); err != nil {
        if log != nil {
            log.Error("CLI execution error", zap.Error(err))
        } else {
            println("❌ CLI execution error:", err.Error()) // Fallback logging
        }
    }
}
