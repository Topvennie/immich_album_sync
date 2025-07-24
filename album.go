package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Album struct {
	ID     string  `json:"id"`
	Name   string  `json:"albumName"`
	Assets []Asset `json:"assets"`
}

func (u *User) getImmichAlbums(ctx context.Context) ([]Album, error) {
	data, err := u.immichRequest(ctx, "GET", "albums", nil)
	if err != nil {
		return nil, err
	}

	var albums []Album
	if err := json.Unmarshal(data, &albums); err != nil {
		return nil, err
	}

	return albums, nil
}

func (u *User) getExternalAlbums() ([]Album, error) {
	albumMap := make(map[string]Album)

	for _, absolutePath := range u.Paths {
		if err := filepath.WalkDir(absolutePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relativePath := path[len(absolutePath):]

			if strings.Count(relativePath, string(os.PathSeparator)) == 1 {
				// Handles album names and top level images
				if d.IsDir() {
					albumMap[d.Name()] = Album{
						Name:   d.Name(),
						Assets: []Asset{},
					}
				}

				return nil
			}

			if d.IsDir() {
				return nil
			}

			parts := strings.Split(relativePath, "/")
			if len(parts) < 2 {
				return nil
			}
			albumName := parts[1]

			album, ok := albumMap[albumName]
			if !ok {
				return nil
			}

			album.Assets = append(albumMap[albumName].Assets, Asset{
				Name: path,
			})
			albumMap[albumName] = album

			return nil
		}); err != nil {
			return nil, err
		}
	}

	albums := make([]Album, 0, len(albumMap))
	for _, album := range albumMap {
		albums = append(albums, album)
	}

	return albums, nil
}

func (u *User) createAlbum(ctx context.Context, album *Album) error {
	fmt.Printf("Creating album %s\n", album.Name)

	type body struct {
		Name string `json:"albumName"`
	}

	data := body{
		Name: album.Name,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := u.immichRequest(ctx, "POST", "albums", bytes.NewReader(buf))
	if err != nil {
		return err
	}

	var newAlbum Album
	if err := json.Unmarshal(resp, &newAlbum); err != nil {
		return err
	}

	album.ID = newAlbum.ID

	return nil
}

func (u *User) syncAlbums(ctx context.Context, external, immich []Album) error {
	immichMap := make(map[string]Album, len(immich))
	for _, album := range immich {
		immichMap[album.Name] = album
	}

	for _, e := range external {
		if _, ok := immichMap[e.Name]; !ok {
			if err := u.createAlbum(ctx, &e); err != nil {
				return err
			}
		}
	}

	return nil
}
