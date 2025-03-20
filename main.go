package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

type Item struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Date  string  `json:"date"`
	Year  string  `json:"year"`
	Month int     `json:"month"`
}

// Item Name - Year - Month - Item
type ItemList map[string]map[string]map[int]Item

var AllItems ItemList

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	AllItems = LoadItems()

	app := fiber.New()

	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("noCache") == "true"
		},
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(AllItems)
	})

	app.Get("/price/:item", func(c *fiber.Ctx) error {
		item := c.Params("item", "eggs")
		return c.JSON(AllItems[item])
	})

	app.Get("/price/:item/:year", func(c *fiber.Ctx) error {
		today := time.Now()
		item := c.Params("item", "eggs")
		year := c.Params("year", strconv.Itoa(today.Year()))
		return c.JSON(AllItems[item][year])
	})

	app.Get("/price/:item/:year/:month", func(c *fiber.Ctx) error {
		today := time.Now()
		item := c.Params("item", "eggs")
		year := c.Params("year", strconv.Itoa(today.Year()))
		month, _ := strconv.Atoi(c.Params("month", strconv.Itoa(int(today.Month()))))
		return c.JSON(AllItems[item][year][month])
	})

	app.Listen(":" + port)
}

func ReadDir(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return []string{}
	}

	fileNames := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			name := strings.Split(entry.Name(), ".")[0]
			fileNames = append(fileNames, name)
		}
	}

	return fileNames
}

func LoadItems() ItemList {
	names := ReadDir("./data/csv")
	allItems := ItemList{}

	for _, i := range names {
		items := ParseItemCSV(i, "./data/csv/"+i+".csv")
		list := make(map[string]map[int]Item)

		for _, item := range items {
			if list[item.Year] == nil {
				list[item.Year] = make(map[int]Item)
			}
			list[item.Year][item.Month] = item
		}

		allItems[i] = list
	}

	return allItems
}

func ParseItemCSV(itemName, path string) []Item {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return nil
	}

	items := []Item{}

	for _, record := range records {
		price, _ := strconv.ParseFloat(record[4], 64)
		month, _ := parseMonth(record[2])

		i := Item{
			Name:  itemName,
			Date:  record[3],
			Value: price,
			Year:  record[1],
			Month: month,
		}

		items = append(items, i)
	}

	return items
}

func parseMonth(monthStr string) (int, error) {
	monthNumStr := strings.TrimPrefix(monthStr, "M")
	monthNum, err := strconv.Atoi(monthNumStr)
	if err != nil {
		return 0, fmt.Errorf("invalid month format: %v", err)
	}
	return monthNum, nil
}
