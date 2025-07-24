package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

type User struct {
	ImmichURL string
	APIKey    string
	ID        string
	Paths     []string
}

func main() {
	config := loadConfig()

	users := make([]User, 0, len(config.Users))

	for _, u := range config.Users {
		user := User{
			ImmichURL: config.ImmichURL,
			APIKey:    u.APIKey,
			Paths:     u.Paths,
		}

		id, err := user.getID(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		user.ID = id

		users = append(users, user)
	}

	ticker := time.NewTicker(time.Duration(config.PollInterval * int(time.Minute)))
	defer ticker.Stop()

	for ; ; <-ticker.C {
		for _, u := range users {
			if err := u.sync(context.Background()); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (u *User) sync(ctx context.Context) error {
	fmt.Println("Syncing albums")

	immich, err := u.getImmichAlbums(ctx)
	if err != nil {
		return err
	}

	external, err := u.getExternalAlbums()
	if err != nil {
		return err
	}

	if err := u.syncAlbums(ctx, external, immich); err != nil {
		return err
	}

	// Refresh immich albums
	immich, err = u.getImmichAlbums(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Syncing album assets")

	for _, album := range immich {
		if err := u.getAlbumAssets(ctx, &album); err != nil {
			return err
		}
	}

	assets, err := u.getImmichAssets(context.Background())
	if err != nil {
		return err
	}

	if err := u.syncAssets(context.Background(), assets, external, immich); err != nil {
		return err
	}

	fmt.Println("Done")

	return nil
}
