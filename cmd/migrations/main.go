package main

import (
	"github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/storage"
)

func main() {
	storage.RunMigration(config.MustNewConfig())
}
