package main

import "github.com/joho/godotenv"

func loadEnv() error {
	return godotenv.Load(".env")
}
