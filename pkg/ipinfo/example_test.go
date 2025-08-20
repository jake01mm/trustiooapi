package ipinfo_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"trusioo_api/pkg/ipinfo"
)

func ExampleNewClient() {
	config := &ipinfo.Config{
		Token:       "your-token-here",
		Timeout:     10 * time.Second,
		CacheEnable: true,
		CacheTTL:    30 * time.Minute,
	}
	
	client := ipinfo.NewClient(config)
	defer client.Close()
	
	ctx := context.Background()
	info, err := client.GetIPInfo(ctx, "8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("IP: %s, City: %s, Country: %s\n", info.IP, info.City, info.Country)
}

func ExampleClient_GetMyIP() {
	config := ipinfo.LoadConfigFromEnv()
	client := ipinfo.NewClient(config)
	defer client.Close()
	
	ctx := context.Background()
	info, err := client.GetMyIP(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("My IP: %s, Location: %s\n", info.IP, info.Loc)
}

func ExampleClient_BatchGetIPInfo() {
	config := &ipinfo.Config{
		Token:       "your-token-here",
		CacheEnable: true,
	}
	
	client := ipinfo.NewClient(config)
	defer client.Close()
	
	ctx := context.Background()
	ips := []string{"8.8.8.8", "1.1.1.1", "208.67.222.222"}
	
	results, err := client.BatchGetIPInfo(ctx, ips)
	if err != nil {
		log.Fatal(err)
	}
	
	for ip, info := range results {
		fmt.Printf("IP: %s, Org: %s\n", ip, info.Org)
	}
}