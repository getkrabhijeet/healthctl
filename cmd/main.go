package main

import (
	"fmt"
	"healthctl/pkg/k8s"
)

func main() {
	client, err := k8s.CreateK8sClientSet()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Client created")

	k8s.GetClusterhealth(client)

	// rook.GetRookClusterHealth(client)

	// mongo.GetMongoClusterHealth(client)

	// redis.GetRedisClusterHealth(client)

	// application.GetApplicationHealth(client)

	k8s.GetAPIResources(client)

}
